package kafkasubscription

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"strconv"
)

//
// Reconcile The Dispatcher (Kafka Consumer) For The Specified Subscription
//
func (r *Reconciler) reconcileDispatcher(ctx context.Context, subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel, responseChannel chan error) {

	// Get Subscription Specific Logger
	logger := util.SubscriptionLogger(r.logger, subscription)

	// If The Subscription Is Being Deleted - Nothing To Do Since K8S Garbage Collection Will Cleanup Based On OwnerReference
	if subscription.DeletionTimestamp != nil {
		logger.Info("Successfully Reconciled Dispatcher Deletion")
		responseChannel <- nil
		return
	}

	// Reconcile The Dispatcher's Service (K8s)
	_, serviceErr := r.createK8sDispatcherService(ctx, subscription, channel)
	if serviceErr != nil {
		r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherK8sServiceReconciliationFailed.String(), "Failed To Reconcile K8S Service For Dispatcher: %v", serviceErr)
		logger.Error("Failed To Reconcile Dispatcher Service", zap.Error(serviceErr))
	} else {
		logger.Info("Successfully Reconciled Dispatcher Service")
	}

	// Reconcile The Dispatcher's Deployment (K8s)
	_, deploymentErr := r.createK8sDispatcherDeployment(ctx, subscription, channel)
	if deploymentErr != nil {
		r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", deploymentErr)
		logger.Error("Failed To Reconcile Dispatcher Deployment", zap.Error(deploymentErr))
	} else {
		logger.Info("Successfully Reconciled Dispatcher Deployment")
	}

	// Return Results
	if serviceErr != nil || deploymentErr != nil {
		responseChannel <- fmt.Errorf("failed to reconcile dispatcher components")
	} else {
		responseChannel <- nil
	}
}

//
// K8S Service
//

// Create The K8S Dispatcher Service If Not Already Existing
func (r *Reconciler) createK8sDispatcherService(ctx context.Context, subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Attempt To Get The K8S Service Associated With The Specified Subscription
	service, err := r.getK8sDispatcherService(ctx, subscription)

	// If The K8S Service Was Not Found - Then Create A New One For The Dispatcher
	if errors.IsNotFound(err) {
		r.logger.Info("Kubernetes Dispatcher Service Not Found - Creating New One")
		service = r.newK8sDispatcherService(subscription, channel)
		err = r.client.Create(ctx, service)
	}

	// If Any Error Occurred (Either Get Or Create) - Then Reconcile Again
	if err != nil {
		return nil, err
	}

	// Return The K8S Service
	return service, nil
}

// Get The K8S Dispatcher Service Associated With The Specified Subscription
func (r *Reconciler) getK8sDispatcherService(ctx context.Context, subscription *messagingv1alpha1.Subscription) (*corev1.Service, error) {

	// Create A Namespace / Name ObjectKey For The Specified Subscription
	serviceKey := types.NamespacedName{
		Namespace: subscription.Namespace,
		Name:      util.DispatcherServiceName(subscription),
	}

	// Get The Service By Namespace / Name
	service := &corev1.Service{}
	err := r.client.Get(ctx, serviceKey, service)

	// Return The Results
	return service, err
}

// Create K8S Dispatcher Service Model For The Specified Subscription
func (r *Reconciler) newK8sDispatcherService(subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel) *corev1.Service {

	// Create & Return The K8S Service Model
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.DispatcherServiceName(subscription),
			Namespace: subscription.Namespace,
			Labels: map[string]string{
				"channel":                     channel.Name,
				DispatcherLabel:               "true",                        // The dispatcher/channel values allows for identification of a Channel's Dispatcher Deployments
				K8sAppDispatcherSelectorLabel: K8sAppDispatcherSelectorValue, // Prometheus ServiceMonitor (See Helm Chart)
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewSubscriptionControllerRef(subscription),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       MetricsPortName,
					Port:       int32(r.environment.MetricsPort),
					TargetPort: intstr.FromInt(r.environment.MetricsPort),
				},
			},
			Selector: map[string]string{
				"app": util.DispatcherDeploymentName(subscription), // Matches Deployment Label Key/Value
			},
		},
	}
}

//
// K8S Deployment
//

// Create The K8S Dispatcher Deployment If Not Already Existing
func (r *Reconciler) createK8sDispatcherDeployment(ctx context.Context, subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Attempt To Get The K8S Dispatcher Deployment Associated With The Specified Channel Subscriber
	deployment, err := r.getK8sDispatcherDeployment(ctx, subscription)

	// If The K8S Dispatcher Deployment Was Not Found - Then Create A New One For The Channel Subscriber
	if errors.IsNotFound(err) {
		r.logger.Info("Kubernetes Dispatcher Deployment Not Found - Creating New One")
		deployment, err = r.newK8sDispatcherDeployment(subscription, channel)
		if err == nil {
			err = r.client.Create(ctx, deployment)
		}
	}

	// If Any Error Occurred (Either Get Or Create) - Then Reconcile Again
	if err != nil {
		return nil, err
	}

	// Return The K8S Dispatcher Deployment
	return deployment, nil
}

// Get The K8S Dispatcher Deployment Associated With The Specified Channel Subscriber
func (r *Reconciler) getK8sDispatcherDeployment(ctx context.Context, subscription *messagingv1alpha1.Subscription) (*appsv1.Deployment, error) {

	// Get The Dispatcher Deployment Name For The Subscription
	deploymentName := util.DispatcherDeploymentName(subscription)

	// Create A Namespace / Name ObjectKey For The Specified Channel Subscriber Deployment
	deploymentKey := types.NamespacedName{
		Namespace: subscription.Namespace,
		Name:      deploymentName,
	}

	// Get The Dispatcher Deployment By Namespace / Name
	deployment := &appsv1.Deployment{}
	err := r.client.Get(ctx, deploymentKey, deployment)

	// Return The Results
	return deployment, err
}

// Create K8S Dispatcher Deployment Model For The Specified Channel Subscriber
func (r *Reconciler) newK8sDispatcherDeployment(subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Get The Dispatcher Deployment Name For The Subscriber
	deploymentName := util.DispatcherDeploymentName(subscription)

	// Replicas Int Value For De-Referencing
	replicas := int32(1)

	// Create The Dispatcher Container Environment Variables
	envVars, err := r.dispatcherDeploymentEnvVars(subscription, channel)
	if err != nil {
		r.logger.Error("Failed To Create Dispatcher Deployment Environment Variables", zap.Error(err))
		return nil, err
	}

	// Create The Dispatcher's K8S Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: channel.Namespace,
			Labels: map[string]string{
				"app":           deploymentName, // Matches K8S Service Selector Key/Value Below
				DispatcherLabel: "true",         // The dispatcher/channel values allows for identification of a Channel's Dispatcher Deployments
				ChannelLabel:    channel.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewSubscriptionControllerRef(subscription),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName, // Matches Template ObjectMeta Pods
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName, // Matched By Deployment Selector Above
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            deploymentName,
							Image:           r.environment.DispatcherImage,
							Env:             envVars,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      constants.LoggingConfigVolumeName,
									MountPath: constants.LoggingConfigMountPath,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: r.environment.DispatcherMemoryLimit,
									corev1.ResourceCPU:    r.environment.DispatcherCpuLimit,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: r.environment.DispatcherMemoryRequest,
									corev1.ResourceCPU:    r.environment.DispatcherCpuRequest,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: constants.LoggingConfigVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: constants.LoggingConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Append EventStartTime Annotation If Populated
	eventStartTime := util.EventStartTime(subscription, r.logger)
	if len(eventStartTime) > 0 {
		deployment.ObjectMeta.Annotations = map[string]string{EventStartTimeAnnotation: eventStartTime}
	}

	// Return The Dispatcher's K8S Deployment
	return deployment, nil
}

// Create The Dispatcher Container's Env Vars
func (r *Reconciler) dispatcherDeploymentEnvVars(subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel) ([]corev1.EnvVar, error) {

	// Get The TopicName For Specified Channel
	topicName := util.TopicName(channel, r.environment)

	// Determine The Subscription's Subscriber URI (Temporary Fallback To Deprecated DNSName Until Kyma Updates To URI)
	subscriberURIRef := subscription.Spec.Subscriber.URI
	if subscriberURIRef == nil || len(subscriberURIRef.String()) <= 0 {
		r.logger.Error("Knative Subscription With Invalid Subscriber Spec - URI Not Specified!")
		return nil, fmt.Errorf("subscription contains invalid subscriber spec -  URI not specified")
	}

	// Create The Dispatcher Deployment EnvVars
	envVars := []corev1.EnvVar{
		{
			Name:  env.MetricsPortEnvVarKey,
			Value: strconv.Itoa(r.environment.MetricsPort),
		},
		{
			Name:  env.KafkaGroupIdEnvVarKey,
			Value: subscription.Name,
		},
		{
			Name:  env.KafkaTopicEnvVarKey,
			Value: topicName,
		},
		{
			Name:  env.KafkaConsumersEnvVarKey,
			Value: strconv.Itoa(util.KafkaConsumers(subscription, r.environment, r.logger)),
		},
		{
			Name:  env.KafkaOffsetCommitMessageCountEnvVarKey,
			Value: strconv.FormatInt(r.environment.KafkaOffsetCommitMessageCount, 10),
		},
		{
			Name:  env.KafkaOffsetCommitDurationMillisEnvVarKey,
			Value: strconv.FormatInt(r.environment.KafkaOffsetCommitDurationMillis, 10),
		},
		{
			Name:  env.SubscriberUriEnvVarKey,
			Value: subscriberURIRef.String(),
		},
		{
			Name:  env.ExponentialBackoffEnvVarKey,
			Value: strconv.FormatBool(util.ExponentialBackoff(subscription, r.environment, r.logger)),
		},
		{
			Name:  env.InitialRetryIntervalEnvVarKey,
			Value: strconv.FormatInt(util.EventRetryInitialIntervalMillis(subscription, r.environment, r.logger), 10),
		},
		{
			Name:  env.MaxRetryTimeEnvVarKey,
			Value: strconv.FormatInt(util.EventRetryTimeMillisMax(subscription, r.environment, r.logger), 10),
		},
	}

	// Get The Kafka Secret From The Kafka Admin Client
	kafkaSecret := r.adminClient.GetKafkaSecretName(topicName)

	// If The Kafka Secret Env Var Is Specified Then Append Relevant Env Vars
	if len(kafkaSecret) <= 0 {

		// Received Invalid Kafka Secret - Cannot Proceed
		return nil, fmt.Errorf("invalid kafkaSecret for topic '%s'", topicName)

	} else {

		// Append The Kafka Brokers As Env Var
		envVars = append(envVars, corev1.EnvVar{
			Name: env.KafkaBrokerEnvVarKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecret},
					Key:                  constants.KafkaSecretDataKeyBrokers,
				},
			},
		})

		// Append The Kafka Username As Env Var
		envVars = append(envVars, corev1.EnvVar{
			Name: env.KafkaUsernameEnvVarKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecret},
					Key:                  constants.KafkaSecretDataKeyUsername,
				},
			},
		})

		// Append The Kafka Password As Env Var
		envVars = append(envVars, corev1.EnvVar{
			Name: env.KafkaPasswordEnvVarKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: kafkaSecret},
					Key:                  constants.KafkaSecretDataKeyPassword,
				},
			},
		})
	}

	// Return The Channel Deployment EnvVars Array
	return envVars, nil
}
