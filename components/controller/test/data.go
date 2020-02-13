package test

import (
	"fmt"
	kafkautil "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	eventingduckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	eventingNames "knative.dev/eventing/pkg/reconciler/names"
	"knative.dev/pkg/apis"
	apisv1 "knative.dev/pkg/apis/duck/v1"
	"strconv"
)

const (
	// Prometheus MetricsPort
	MetricsPortName = "metrics"

	// Controller Config Test Data
	ServiceAccount                         = "TestServiceAccount"
	MetricsPort                            = 9876
	KafkaSecret                            = "testkafkasecret"
	KafkaOffsetCommitMessageCount          = 99
	KafkaOffsetCommitDurationMillis        = 9999
	ChannelImage                           = "TestChannelImage"
	ChannelReplicas                        = 1
	DispatcherImage                        = "TestDispatcherImage"
	DispatcherReplicas                     = 1
	DefaultNumPartitions                   = 4
	DefaultReplicationFactor               = 1
	DefaultRetentionMillis                 = 99999
	DefaultEventRetryInitialIntervalMillis = 88888
	DefaultEventRetryTimeMillisMax         = 11111111
	DefaultExponentialBackoff              = true
	DefaultDispatcherReplicas              = 1

	// Channel Test Data
	TenantId                 = "TestTenantId"
	ChannelName              = "TestChannelName"
	NamespaceName            = "TestNamespaceName"
	ChannelServiceName       = "testchannelname-testnamespacename-channel"
	ChannelDeploymentName    = KafkaSecret + "-channel"
	SubscriberName           = "test-subscriber-name"
	DispatcherDeploymentName = "testchannelname-testnamespacename-dispatcher"
	DefaultTopicName         = NamespaceName + "." + ChannelName
	TopicName                = NamespaceName + "." + ChannelName

	// Subscription Test Data
	EventStartTime = "2019-01-01T00:00:00Z"

	// Channel Arguments Test Data
	NumPartitions     = 123
	ReplicationFactor = 456
	RetentionMillis   = 999999999

	// Test MetaData
	ErrorString = "Expected Mock Test Error"

	// Test Dispatcher Resources
	DispatcherMemoryRequest = "20Mi"
	DispatcherCpuRequest    = "100m"
	DispatcherMemoryLimit   = "50Mi"
	DispatcherCpuLimit      = "300m"

	// Test Channel Resources
	ChannelMemoryRequest = "10Mi"
	ChannelMemoryLimit   = "20Mi"
	ChannelCpuRequest    = "10m"
	ChannelCpuLimit      = "100m"
)

var (
	// Subscriber URI (Because We Can't Get The Address Of Constants!)
	SubscriberURI = "TestSubscriberURI"

	// Constant For Deleted Timestamps
	DeletedTimestamp = metav1.Now().Rfc3339Copy()
)

//
// ControllerConfig Test Data
//

// Set The Required Environment Variables
func NewEnvironment() *env.Environment {
	return &env.Environment{
		ServiceAccount:                       ServiceAccount,
		MetricsPort:                          MetricsPort,
		KafkaOffsetCommitMessageCount:        KafkaOffsetCommitMessageCount,
		KafkaOffsetCommitDurationMillis:      KafkaOffsetCommitDurationMillis,
		DefaultNumPartitions:                 DefaultNumPartitions,
		DefaultReplicationFactor:             DefaultReplicationFactor,
		DefaultRetentionMillis:               DefaultRetentionMillis,
		DispatcherImage:                      DispatcherImage,
		DispatcherReplicas:                   DispatcherReplicas,
		DispatcherRetryInitialIntervalMillis: DefaultEventRetryInitialIntervalMillis,
		DispatcherRetryTimeMillisMax:         DefaultEventRetryTimeMillisMax,
		DispatcherRetryExponentialBackoff:    DefaultExponentialBackoff,
		DispatcherCpuLimit:                   resource.MustParse(DispatcherCpuLimit),
		DispatcherCpuRequest:                 resource.MustParse(DispatcherCpuRequest),
		DispatcherMemoryLimit:                resource.MustParse(DispatcherMemoryLimit),
		DispatcherMemoryRequest:              resource.MustParse(DispatcherMemoryRequest),
		ChannelImage:                         ChannelImage,
		ChannelReplicas:                      ChannelReplicas,
		ChannelMemoryRequest:                 resource.MustParse(ChannelMemoryRequest),
		ChannelMemoryLimit:                   resource.MustParse(ChannelMemoryLimit),
		ChannelCpuRequest:                    resource.MustParse(ChannelCpuRequest),
		ChannelCpuLimit:                      resource.MustParse(ChannelCpuLimit),
	}
}

//
// K8S Test Model Utility Functions
//

// Utility Function For Creating A Test Channel With Specified State
func GetNewChannel(includeFinalizer bool, includeSpecProperties bool, includeSubscribers bool) *kafkav1alpha1.KafkaChannel {

	// Create The Specified Channel
	channel := &kafkav1alpha1.KafkaChannel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kafkav1alpha1.SchemeGroupVersion.String(),
			Kind:       constants.KafkaChannelKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NamespaceName,
			Name:      ChannelName,
		},
	}

	// Append The Finalizer To The MetaData Spec If Specified
	if includeFinalizer {
		channel.ObjectMeta.Finalizers = []string{constants.KafkaChannelControllerAgentName}
	}

	// Include Channel Spec Properties If Specified
	if includeSpecProperties {
		channel.Spec.TenantId = TenantId
		channel.Spec.NumPartitions = NumPartitions
		channel.Spec.ReplicationFactor = ReplicationFactor
		channel.Spec.RetentionMillis = RetentionMillis
	}

	// Append Subscribers To Channel Spec If Specified
	if includeSubscribers {
		url, _ := apis.ParseURL(SubscriberURI)
		channel.Spec.Subscribers = []eventingduckv1beta1.SubscriberSpec{
			{
				UID:           SubscriberName,
				SubscriberURI: url,
			},
		}
	}

	// Return The Channel
	return channel
}

// Utility Function For Creating A Test Channel With Specified Name
func GetNewChannelWithName(name string, includeFinalizer bool, includeSpecProperties bool, includeSubscribers bool) *kafkav1alpha1.KafkaChannel {

	// Get The Default Test Channel
	channel := GetNewChannel(includeFinalizer, includeSpecProperties, includeSubscribers)

	// Customize The Name
	channel.ObjectMeta.Name = name

	// Return The Channel
	return channel
}

// Utility Function For Creating A Test Channel (Provisioned) With Deletion Timestamp
func GetNewChannelDeleted(includeFinalizer bool, includeSpecProperties bool, includeSubscribers bool) *kafkav1alpha1.KafkaChannel {
	channel := GetNewChannelWithProvisionedStatus(includeFinalizer, includeSpecProperties, includeSubscribers, true, GetChannelStatusReady())
	channel.DeletionTimestamp = &DeletedTimestamp
	if !includeFinalizer {
		channel.ObjectMeta.Finalizers = nil
	}
	return channel
}

// Utility Function For Creating A Test Channel With Provisioned / NotProvisioned Status
func GetNewChannelWithProvisionedStatus(includeFinalizer bool, includeSpecProperties bool, includeSubscribers bool, provisioned bool, status *kafkav1alpha1.KafkaChannelStatus) *kafkav1alpha1.KafkaChannel {
	channel := GetNewChannel(includeFinalizer, includeSpecProperties, includeSubscribers)
	if provisioned {
		channel.Status = *status
	}
	return channel
}

// Get READY Success Status
func GetChannelStatusReady() *kafkav1alpha1.KafkaChannelStatus {
	return GetChannelStatus(true, true, true, true, true)
}

func GetChannelStatusTopicFailed() *kafkav1alpha1.KafkaChannelStatus {
	return GetChannelStatus(false, true, true, true, true)
}

func GetChannelStatusChannelServiceFailed() *kafkav1alpha1.KafkaChannelStatus {
	return GetChannelStatus(true, false, true, true, true)
}

func GetChannelStatusChannelDeploymentServiceFailed() *kafkav1alpha1.KafkaChannelStatus {
	return GetChannelStatus(true, true, false, true, true)
}

func GetChannelStatusChannelDeploymentFailed() *kafkav1alpha1.KafkaChannelStatus {
	return GetChannelStatus(true, true, true, false, true)
}

func GetChannelStatusDispatcherDeploymentFailed() *kafkav1alpha1.KafkaChannelStatus {
	return GetChannelStatus(true, true, true, true, false)
}

// Utility Function For Creating A KafkaChannel Status
func GetChannelStatus(topic, channelService, channelDeploymentService, channelDeployment, dispatcherDeployment bool) *kafkav1alpha1.KafkaChannelStatus {
	kafkaChannelStatus := &kafkav1alpha1.KafkaChannelStatus{}
	kafkaChannelStatus.InitializeConditions()

	if topic {
		kafkaChannelStatus.MarkTopicTrue()
	} else {
		kafkaChannelStatus.MarkTopicFailed("TopicFailed", "Channel Kafka Topic Failed: Test Error")
	}

	if channelService {
		kafkaChannelStatus.MarkChannelServiceTrue()
		kafkaChannelStatus.SetAddress(&apis.URL{
			Scheme: "http",
			Host:   eventingNames.ServiceHostName(kafkautil.AppendKafkaChannelServiceNameSuffix(ChannelName), NamespaceName),
		})
	} else {
		kafkaChannelStatus.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service Failed: %s", MockCreateFnServiceErrorMessage))
	}

	if channelDeploymentService {
		kafkaChannelStatus.MarkChannelDeploymentServiceTrue()
	} else {
		kafkaChannelStatus.MarkChannelDeploymentServiceFailed("ChannelDeploymentServiceFailed", fmt.Sprintf("Channel Deployment Service Failed: %s", MockCreateFnServiceErrorMessage))
	}

	if channelDeployment {
		kafkaChannelStatus.MarkChannelDeploymentTrue()
	} else {
		kafkaChannelStatus.MarkChannelDeploymentFailed("ChannelDeploymentFailed", fmt.Sprintf("Channel Deployment Failed: %s", MockCreateFnDeploymentErrorMessage))
	}

	if dispatcherDeployment {
		kafkaChannelStatus.MarkDispatcherDeploymentTrue()
	} else {
		kafkaChannelStatus.MarkDispatcherDeploymentFailed("DispatcherDeploymentFailed", fmt.Sprintf("Dispatcher Deployment Failed: %s", MockCreateFnDeploymentErrorMessage))
	}

	return kafkaChannelStatus
}

// Utility Function For Creating A K8S Channel Service For The Test Channel
func GetNewKafkaChannelService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkautil.AppendKafkaChannelServiceNameSuffix(ChannelName),
			Namespace: NamespaceName,
			Labels: map[string]string{
				"k8s-app": "knative-kafka-channels",
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewChannelOwnerRef(true),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: ChannelDeploymentName + "." + constants.KnativeEventingNamespace + ".svc.cluster.local",
		},
	}
}

// Utility Function For Creating A K8S Channel Service For The Test Channel
func GetNewChannelDeploymentService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ChannelDeploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"k8s-app": "knative-kafka-channels",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       constants.HttpPortName,
					Port:       constants.HttpServicePortNumber,
					TargetPort: intstr.FromInt(constants.HttpContainerPortNumber),
				},
				{
					Name:       MetricsPortName,
					Port:       MetricsPort,
					TargetPort: intstr.FromInt(MetricsPort),
				},
			},
			Selector: map[string]string{
				"app": ChannelDeploymentName,
			},
		},
	}
}

// Utility Function For Creating A K8S Channel Deployment For The Test Channel
func GetNewK8SChannelDeployment(topicName string) *appsv1.Deployment {
	replicas := int32(ChannelReplicas)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ChannelDeploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"app": ChannelDeploymentName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": ChannelDeploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": ChannelDeploymentName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:  ChannelDeploymentName,
							Image: ChannelImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "server",
									ContainerPort: int32(8080),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "METRICS_PORT",
									Value: strconv.Itoa(MetricsPort),
								},
								{
									Name: env.KafkaBrokerEnvVarKey,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: KafkaSecret},
											Key:                  constants.KafkaSecretDataKeyBrokers,
										},
									},
								},
								{
									Name: env.KafkaUsernameEnvVarKey,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: KafkaSecret},
											Key:                  constants.KafkaSecretDataKeyUsername,
										},
									},
								},
								{
									Name: env.KafkaPasswordEnvVarKey,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: KafkaSecret},
											Key:                  constants.KafkaSecretDataKeyPassword,
										},
									},
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      constants.LoggingConfigVolumeName,
									MountPath: constants.LoggingConfigMountPath,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(ChannelCpuRequest),
									corev1.ResourceMemory: resource.MustParse(ChannelMemoryRequest),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(ChannelCpuLimit),
									corev1.ResourceMemory: resource.MustParse(ChannelMemoryLimit),
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
}

// Utility Function For Creating A Test Subscription With Specified State
func GetNewSubscription(namespaceName string, subscriberName string, includeAnnotations bool, includeFinalizer bool, eventStartTime string) *messagingv1beta1.Subscription {

	// Parse The Subscriber URI String Into A URI
	subscriberURI, _ := apis.ParseURL(SubscriberURI)

	// Create The Specified Subscription
	subscription := &messagingv1beta1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: messagingv1beta1.SchemeGroupVersion.String(),
			Kind:       constants.KnativeSubscriptionKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespaceName,
			Name:      subscriberName,
		},
		Spec: messagingv1beta1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				APIVersion: kafkav1alpha1.SchemeGroupVersion.String(),
				Kind:       constants.KafkaChannelKind,
				Name:       ChannelName,
			},
			Subscriber: &apisv1.Destination{
				URI: subscriberURI,
			},
		},
	}

	// Set The Annotations If Specified
	if includeAnnotations {
		subscription.ObjectMeta.Annotations = map[string]string{
			"knativekafka.kyma-project.io/EventRetryInitialIntervalMillis": strconv.Itoa(DefaultEventRetryInitialIntervalMillis),
			"knativekafka.kyma-project.io/EventRetryTimeMillisMax":         strconv.Itoa(DefaultEventRetryTimeMillisMax),
			"knativekafka.kyma-project.io/ExponentialBackoff":              strconv.FormatBool(DefaultExponentialBackoff),
			"knativekafka.kyma-project.io/EventStartTime":                  eventStartTime,
		}
	}

	// Set The Finalizer To The MetaData Spec If Specified
	if includeFinalizer {
		subscription.ObjectMeta.Finalizers = []string{constants.KafkaSubscriptionControllerAgentName}
	}

	// Return The Subscription
	return subscription
}

// Utility Function For Creating A Test Subscription (Provisioned) With Deletion Timestamp
func GetNewSubscriptionDeleted(namespaceName string, subscriberName string, includeAnnotations bool, includeFinalizer bool) *messagingv1beta1.Subscription {
	subscription := GetNewSubscription(namespaceName, subscriberName, includeAnnotations, includeFinalizer, EventStartTime)
	subscription.DeletionTimestamp = &DeletedTimestamp
	return subscription
}

// Utility Function For Creating A K8S Dispatcher Service For The Test Channel
func GetNewK8SDispatcherService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DispatcherDeploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"channel":    ChannelName,
				"dispatcher": "true",
				"k8s-app":    "knative-kafka-dispatchers",
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewChannelOwnerRef(true),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       MetricsPortName,
					Port:       int32(MetricsPort),
					TargetPort: intstr.FromInt(MetricsPort),
				},
			},
			Selector: map[string]string{
				"app": DispatcherDeploymentName,
			},
		},
	}
}

// Utility Function For Creating A K8S Dispatcher Deployment For The Test Channel
func GetNewK8SDispatcherDeployment(topicName string) *appsv1.Deployment {
	replicas := int32(DispatcherReplicas)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DispatcherDeploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"app":        DispatcherDeploymentName,
				"dispatcher": "true",
				"channel":    ChannelName,
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewChannelOwnerRef(true),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": DispatcherDeploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": DispatcherDeploymentName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:  DispatcherDeploymentName,
							Image: DispatcherImage,
							Env: []corev1.EnvVar{
								{
									Name:  env.MetricsPortEnvVarKey,
									Value: strconv.Itoa(MetricsPort),
								},
								{
									Name:  env.ChannelKeyEnvVarKey,
									Value: fmt.Sprintf("%s/%s", NamespaceName, ChannelName),
								},
								{
									Name:  env.KafkaTopicEnvVarKey,
									Value: topicName,
								},
								{
									Name:  env.KafkaOffsetCommitMessageCountEnvVarKey,
									Value: strconv.Itoa(KafkaOffsetCommitMessageCount),
								},
								{
									Name:  env.KafkaOffsetCommitDurationMillisEnvVarKey,
									Value: strconv.Itoa(KafkaOffsetCommitDurationMillis),
								},
								{
									Name:  env.ExponentialBackoffEnvVarKey,
									Value: strconv.FormatBool(DefaultExponentialBackoff),
								},
								{
									Name:  env.InitialRetryIntervalEnvVarKey,
									Value: strconv.Itoa(DefaultEventRetryInitialIntervalMillis),
								},
								{
									Name:  env.MaxRetryTimeEnvVarKey,
									Value: strconv.Itoa(DefaultEventRetryTimeMillisMax),
								},
								{
									Name: env.KafkaBrokerEnvVarKey,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: KafkaSecret},
											Key:                  constants.KafkaSecretDataKeyBrokers,
										},
									},
								},
								{
									Name: env.KafkaUsernameEnvVarKey,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: KafkaSecret},
											Key:                  constants.KafkaSecretDataKeyUsername,
										},
									},
								},
								{
									Name: env.KafkaPasswordEnvVarKey,
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{Name: KafkaSecret},
											Key:                  constants.KafkaSecretDataKeyPassword,
										},
									},
								},
							},
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      constants.LoggingConfigVolumeName,
									MountPath: constants.LoggingConfigMountPath,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse(DispatcherMemoryLimit),
									corev1.ResourceCPU:    resource.MustParse(DispatcherCpuLimit),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse(DispatcherMemoryRequest),
									corev1.ResourceCPU:    resource.MustParse(DispatcherCpuRequest),
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
}

// Create A New Dispatcher Deployment And Clear The Annotations
func GetNewK8SDispatcherDeploymentWithoutAnnotations(topicName string) *appsv1.Deployment {
	deployment := GetNewK8SDispatcherDeployment(topicName)
	deployment.ObjectMeta.Annotations = nil
	return deployment

}

// Create A New OwnerReference Model For The Test Channel
func GetNewChannelOwnerRef(isController bool) metav1.OwnerReference {
	blockOwnerDeletion := true
	return metav1.OwnerReference{
		APIVersion:         "knativekafka.kyma-project.io/v1alpha1",
		Kind:               constants.KafkaChannelKind,
		Name:               ChannelName,
		UID:                "",
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}
