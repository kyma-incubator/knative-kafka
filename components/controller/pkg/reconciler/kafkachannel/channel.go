package kafkachannel

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
	eventingNames "knative.dev/eventing/pkg/reconciler/names"
	"knative.dev/pkg/apis"
	"strconv"
)

// Reconcile The "Channel" Inbound For The Specified Channel (K8S Service, Istio VirtualService, Deployment)
func (r *Reconciler) reconcileChannel(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// Get Channel Specific Logger
	logger := util.ChannelLogger(r.logger, channel)

	// If The Channel Is Being Deleted - Nothing To Do Since K8S Garbage Collection Will Cleanup Based On OwnerReference
	if channel.DeletionTimestamp != nil {
		logger.Info("Successfully Reconciled Channel Deletion")
		return nil
	}

	// TODO Note, similar to the topic implementation there is not yet any intelligent handling of
	//      changes to the Controller config or Channel arguments - need requirements for such ; )

	// Reconcile The Channel's Service (K8s)
	_, serviceErr := r.createK8sChannelService(ctx, channel)
	if serviceErr != nil {
		r.recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile K8S Service For Channel: %v", serviceErr)
		logger.Error("Failed To Reconcile Channel Service", zap.Error(serviceErr))
	} else {
		logger.Info("Successfully Reconciled Channel Service")
	}

	// Reconcile The Channel's Deployment (K8s)
	_, deploymentErr := r.createK8sChannelDeployment(ctx, channel)
	if deploymentErr != nil {
		r.recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelDeploymentReconciliationFailed.String(), "Failed To Reconcile Deployment For Channel: %v", deploymentErr)
		logger.Error("Failed To Reconcile Channel Deployment", zap.Error(deploymentErr))
	} else {
		logger.Info("Successfully Reconciled Channel Deployment")
	}

	// Return Results
	if serviceErr != nil || deploymentErr != nil {
		return fmt.Errorf("failed to reconcile channel components")
	} else {
		return nil
	}
}

//
// K8S Channel Service
//
// Note - These functions were lifted from early Knative Eventing implementation which have since been moved/removed. // TODO - review this paragraph and remove/reword
//        Ideally we'd like to use that implementation directly, but it uses hardcoded config for the larger Knative Eventing
//        ClusterBus / Single Dispatcher paradigm which doesn't work for our scalable implementation where we want a unique
//        service/channel deployment per channel/topic.  The logic has also been modified for readability but is otherwise similar.
//

// Create The K8S Service If Not Already Existing
func (r *Reconciler) createK8sChannelService(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Attempt To Get The K8S Service Associated With The Specified Channel
	service, err := r.getK8sChannelService(ctx, channel)

	// If The K8S Service Was Not Found - Then Create A New One For The Channel
	if errors.IsNotFound(err) {
		r.logger.Info("Kubernetes Channel Service Not Found - Creating New One")
		service = r.newK8sChannelService(channel)
		err = r.client.Create(ctx, service)
	}

	// If Any Error Occurred (Either Get Or Create) - Then Reconcile Again
	if err != nil {
		channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service Failed: %s", err))
		return nil, err
	}

	//// Update The Channel Status With Service Address
	channel.Status.MarkChannelServiceTrue()
	channel.Status.SetAddress(&apis.URL{
		Scheme: "http",
		Host:   eventingNames.ServiceHostName(service.Name, service.Namespace),
	})

	// Return The K8S Service
	return service, nil
}

// Get The K8S Channel Service Associated With The Specified Channel
func (r *Reconciler) getK8sChannelService(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Create A Namespace / Name ObjectKey For The Specified Channel
	serviceKey := types.NamespacedName{
		Namespace: channel.Namespace,
		Name:      util.ChannelServiceName(channel.Name),
	}

	// Get The Service By Namespace / Name
	service := &corev1.Service{}
	err := r.client.Get(ctx, serviceKey, service)

	// Return The Results
	return service, err
}

// Create K8S Channel Service Model For The Specified Channel
func (r *Reconciler) newK8sChannelService(channel *kafkav1alpha1.KafkaChannel) *corev1.Service {

	// Create & Return The K8S Service Model
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ChannelServiceName(channel.ObjectMeta.Name),
			Namespace: channel.Namespace,
			Labels: map[string]string{
				"channel":                  channel.Name,
				K8sAppChannelSelectorLabel: K8sAppChannelSelectorValue, // Prometheus ServiceMonitor (See Helm Chart)
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelControllerRef(channel),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       constants.HttpPortName,
					Port:       constants.HttpPortNumber,
					TargetPort: intstr.FromInt(constants.HttpPortNumber),
				},
				{
					Name:       constants.MetricsPortName,
					Port:       int32(r.environment.MetricsPort),
					TargetPort: intstr.FromInt(r.environment.MetricsPort),
				},
			},
			Selector: map[string]string{
				"app": channelDeploymentName(channel), // Matches Deployment Label Key/Value
			},
		},
	}
}

//
// K8S Deployment
//

// Create The K8S Channel Deployment If Not Already Existing
func (r *Reconciler) createK8sChannelDeployment(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Attempt To Get The K8S Channel Deployment Associated With The Specified Channel
	deployment, err := r.getK8sChannelDeployment(ctx, channel)

	// If The K8S Channel Deployment Was Not Found - Then Create A New One For The Channel
	if errors.IsNotFound(err) {
		r.logger.Info("Kubernetes Channel Deployment Not Found - Creating New One")
		deployment, err = r.newK8sChannelDeployment(channel)
		if err == nil {
			err = r.client.Create(ctx, deployment)
		}
	}

	// If Any Error Occurred (Either Get Or Create) - Then Reconcile Again
	if err != nil {
		return nil, err
	}

	// Return The K8S Channel Deployment
	return deployment, nil
}

// Get The K8S Channel Deployment Associated With The Specified Channel
func (r *Reconciler) getK8sChannelDeployment(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Get The Channel Deployment Name
	deploymentName := channelDeploymentName(channel)

	// Create A Namespace / Name ObjectKey For The Specified Channel Deployment
	deploymentKey := types.NamespacedName{
		Namespace: channel.Namespace,
		Name:      deploymentName,
	}

	// Get The Channel Deployment By Namespace / Name
	deployment := &appsv1.Deployment{}
	err := r.client.Get(ctx, deploymentKey, deployment)

	// Return The Results
	return deployment, err
}

// Create K8S Channel Deployment Model For The Specified Channel
func (r *Reconciler) newK8sChannelDeployment(channel *kafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Get The Channel Deployment Name
	deploymentName := channelDeploymentName(channel)

	// Replicas Int Value For De-Referencing
	replicas := int32(1)

	// Create The Channel Container Environment Variables
	channelEnvVars, err := r.channelDeploymentEnvVars(channel)
	if err != nil {
		r.logger.Error("Failed To Create Channel Deployment Environment Variables", zap.Error(err))
		return nil, err
	}

	// Create & Return The Channel's K8S Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: channel.Namespace,
			Labels: map[string]string{
				"app": deploymentName, // Matches K8S Service Selector Key/Value Below
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelControllerRef(channel),
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
							Name:  deploymentName,
							Image: r.environment.ChannelImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "server",
									ContainerPort: int32(constants.HttpPortNumber),
								},
							},
							Env:             channelEnvVars,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      constants.LoggingConfigVolumeName,
									MountPath: constants.LoggingConfigMountPath,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    r.environment.ChannelCpuRequest,
									corev1.ResourceMemory: r.environment.ChannelMemoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    r.environment.ChannelCpuLimit,
									corev1.ResourceMemory: r.environment.ChannelMemoryLimit,
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

	// Return Channel Deployment
	return deployment, nil
}

// Create The Channel Container's Env Vars
func (r *Reconciler) channelDeploymentEnvVars(channel *kafkav1alpha1.KafkaChannel) ([]corev1.EnvVar, error) {

	// Get The TopicName For Specified Channel
	topicName := util.TopicName(channel, r.environment)

	// Create The Channel Deployment EnvVars
	envVars := []corev1.EnvVar{
		{
			Name:  env.HttpPortEnvVarKey,
			Value: strconv.Itoa(constants.HttpPortNumber),
		},
		{
			Name:  env.MetricsPortEnvVarKey,
			Value: strconv.Itoa(r.environment.MetricsPort),
		},
		{
			Name:  env.KafkaTopicEnvVarKey,
			Value: topicName,
		},
		{
			Name:  env.KafkaClientIdEnvVarKey,
			Value: topicName,
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

//
// Utilities
//

// Get The Channel's K8S Deployment Name
func channelDeploymentName(channel *kafkav1alpha1.KafkaChannel) string {
	return fmt.Sprintf("%s-channel", channel.Name)
}
