package kafkachannel

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

//
// Reconcile The Dispatcher (Kafka Consumer) For The Specified KafkaChannel
//
func (r *Reconciler) reconcileDispatcher(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Get Channel Specific Logger
	logger := util.ChannelLogger(r.Logger.Desugar(), channel)

	// If The KafkaChannel Is Being Deleted - Nothing To Do Since K8S Garbage Collection Will Cleanup Based On OwnerReference
	if channel.DeletionTimestamp != nil {
		logger.Info("Successfully Reconciled Dispatcher Deletion")
		return nil
	}

	// Reconcile The Dispatcher's Service (For Prometheus Only)
	_, serviceErr := r.createK8sDispatcherService(channel)
	if serviceErr != nil {
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.DispatcherServiceReconciliationFailed.String(), "Failed To Reconcile K8S Service For Dispatcher: %v", serviceErr)
		logger.Error("Failed To Reconcile Dispatcher Service", zap.Error(serviceErr))
	} else {
		logger.Info("Successfully Reconciled Dispatcher Service")
	}

	// Reconcile The Dispatcher's Deployment
	_, deploymentErr := r.createK8sDispatcherDeployment(channel)
	if deploymentErr != nil {
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", deploymentErr)
		logger.Error("Failed To Reconcile Dispatcher Deployment", zap.Error(deploymentErr))
		channel.Status.MarkDispatcherDeploymentFailed("DispatcherDeploymentFailed", fmt.Sprintf("Dispatcher Deployment Failed: %s", deploymentErr))
	} else {
		logger.Info("Successfully Reconciled Dispatcher Deployment")
		channel.Status.MarkDispatcherDeploymentTrue()
	}

	// Return Results
	if serviceErr != nil || deploymentErr != nil {
		return fmt.Errorf("failed to reconcile dispatcher components")
	} else {
		return nil
	}
}

//
// K8S Service
//

// Create The K8S Dispatcher Service If Not Already Existing
func (r *Reconciler) createK8sDispatcherService(channel *knativekafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Attempt To Get The K8S Service Associated With The Specified Channel
	service, err := r.getK8sDispatcherService(channel)

	// If The K8S Service Was Not Found - Then Create A New One For The Dispatcher
	if errors.IsNotFound(err) {
		r.Logger.Info("Kubernetes Dispatcher Service Not Found - Creating New One")
		service = r.newK8sDispatcherService(channel)
		service, err = r.KubeClientSet.CoreV1().Services(service.Namespace).Create(service)
	}

	// If Any Error Occurred (Either Get Or Create) - Then Reconcile Again
	if err != nil {
		return nil, err
	}

	// Return The K8S Service
	return service, nil
}

// Get The K8S Dispatcher Service Associated With The Specified Channel
func (r *Reconciler) getK8sDispatcherService(channel *knativekafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Get The Service By Namespace / Name
	service := &corev1.Service{}
	service, err := r.serviceLister.Services(constants.KnativeEventingNamespace).Get(util.DispatcherDnsSafeName(channel))

	// Return The Results
	return service, err
}

// Create K8S Dispatcher Service Model For The Specified Subscription
func (r *Reconciler) newK8sDispatcherService(channel *knativekafkav1alpha1.KafkaChannel) *corev1.Service {

	// Get The Dispatcher Service Name For The Channel
	serviceName := util.DispatcherDnsSafeName(channel)

	// Create & Return The K8S Service Model
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"channel":                     channel.Name,
				DispatcherLabel:               "true",                        // The dispatcher/channel values allows for identification of a Channel's Dispatcher Deployments
				K8sAppDispatcherSelectorLabel: K8sAppDispatcherSelectorValue, // Prometheus ServiceMonitor (See Helm Chart)
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelOwnerReference(channel),
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
				"app": serviceName, // Matches Deployment Label Key/Value
			},
		},
	}
}

//
// K8S Deployment
//

// Create The K8S Dispatcher Deployment If Not Already Existing
func (r *Reconciler) createK8sDispatcherDeployment(channel *knativekafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Attempt To Get The K8S Dispatcher Deployment Associated With The Specified Channel
	deployment, err := r.getK8sDispatcherDeployment(channel)

	// If The K8S Dispatcher Deployment Was Not Found - Then Create A New One For The Channel
	if errors.IsNotFound(err) {
		r.Logger.Info("Kubernetes Dispatcher Deployment Not Found - Creating New One")
		deployment, err = r.newK8sDispatcherDeployment(channel)
		if err == nil {
			deployment, err = r.KubeClientSet.AppsV1().Deployments(deployment.Namespace).Create(deployment)
		}
	}

	// If Any Error Occurred (Either Get Or Create) - Then Reconcile Again
	if err != nil {
		return nil, err
	}

	// Return The K8S Dispatcher Deployment
	return deployment, nil
}

// Get The K8S Dispatcher Deployment Associated With The Specified Channel
func (r *Reconciler) getK8sDispatcherDeployment(channel *knativekafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Get The Dispatcher Deployment Name For The Channel
	deploymentName := util.DispatcherDnsSafeName(channel)

	// Get The Dispatcher Deployment By Namespace / Name
	deployment := &appsv1.Deployment{}
	deployment, err := r.deploymentLister.Deployments(constants.KnativeEventingNamespace).Get(deploymentName)

	// Return The Results
	return deployment, err
}

// Create K8S Dispatcher Deployment Model For The Specified Channel
func (r *Reconciler) newK8sDispatcherDeployment(channel *knativekafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Get The Dispatcher Deployment Name For The Channel
	deploymentName := util.DispatcherDnsSafeName(channel)

	// Replicas Int Value For De-Referencing
	replicas := int32(r.environment.DispatcherReplicas)

	// Create The Dispatcher Container Environment Variables
	envVars, err := r.dispatcherDeploymentEnvVars(channel)
	if err != nil {
		r.Logger.Error("Failed To Create Dispatcher Deployment Environment Variables", zap.Error(err))
		return nil, err
	}

	// Create The Dispatcher's K8S Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"app":           deploymentName, // Matches K8S Service Selector Key/Value Below
				DispatcherLabel: "true",         // The dispatcher/channel values allows for identification of a Channel's Dispatcher Deployments
				ChannelLabel:    channel.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelOwnerReference(channel),
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
					ServiceAccountName: r.environment.ServiceAccount,
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

	// Return The Dispatcher's K8S Deployment
	return deployment, nil
}

// Create The Dispatcher Container's Env Vars
func (r *Reconciler) dispatcherDeploymentEnvVars(channel *knativekafkav1alpha1.KafkaChannel) ([]corev1.EnvVar, error) {

	// Get The TopicName For Specified Channel
	topicName := util.TopicName(channel)

	// Create The Dispatcher Deployment EnvVars
	envVars := []corev1.EnvVar{
		{
			Name:  env.MetricsPortEnvVarKey,
			Value: strconv.Itoa(r.environment.MetricsPort),
		},
		{
			Name:  env.ChannelKeyEnvVarKey,
			Value: util.ChannelKey(channel),
		},
		{
			Name:  env.KafkaTopicEnvVarKey,
			Value: topicName,
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
			Name:  env.ExponentialBackoffEnvVarKey,
			Value: strconv.FormatBool(r.environment.DispatcherRetryExponentialBackoff),
		},
		{
			Name:  env.InitialRetryIntervalEnvVarKey,
			Value: strconv.FormatInt(r.environment.DispatcherRetryInitialIntervalMillis, 10),
		},
		{
			Name:  env.MaxRetryTimeEnvVarKey,
			Value: strconv.FormatInt(r.environment.DispatcherRetryTimeMillisMax, 10),
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
