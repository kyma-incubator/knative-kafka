package test

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	eventingduck "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	eventingduckv1alpha1 "knative.dev/eventing/pkg/apis/duck/v1alpha1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	eventingNames "knative.dev/eventing/pkg/reconciler/names"
	"knative.dev/pkg/apis"
	"strconv"
)

const (
	// Prometheus MetricsPort
	MetricsPortName = "metrics"

	// Controller Config Test Data
	MetricsPort                            = 9876
	KafkaSecret                            = "TestKafkaSecret"
	KafkaOffsetCommitMessageCount          = 99
	KafkaOffsetCommitDurationMillis        = 9999
	ChannelImage                           = "TestChannelImage"
	DispatcherImage                        = "TestDispatcherImage"
	DefaultNumPartitions                   = 4
	DefaultReplicationFactor               = 1
	DefaultRetentionMillis                 = 99999
	DefaultEventRetryInitialIntervalMillis = 88888
	DefaultEventRetryTimeMillisMax         = 11111111
	DefaultExponentialBackoff              = true
	DefaultKafkaConsumers                  = 4

	// Channel Test Data
	TenantId                 = "TestTenantId"
	DefaultTenantId          = "TestDefaultTenantId"
	ChannelName              = "TestChannelName"
	NamespaceName            = "TestNamespaceName"
	ChannelDeploymentName    = ChannelName + "-channel"
	SubscriberName           = "test-subscriber-name"
	DispatcherDeploymentName = SubscriberName + "-dispatcher"
	DefaultTopicName         = DefaultTenantId + "." + NamespaceName + "." + ChannelName
	TopicName                = TenantId + "." + NamespaceName + "." + ChannelName

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
		MetricsPort:                            MetricsPort,
		KafkaOffsetCommitMessageCount:          KafkaOffsetCommitMessageCount,
		KafkaOffsetCommitDurationMillis:        KafkaOffsetCommitDurationMillis,
		ChannelImage:                           ChannelImage,
		DispatcherImage:                        DispatcherImage,
		DefaultTenantId:                        DefaultTenantId,
		DefaultNumPartitions:                   DefaultNumPartitions,
		DefaultReplicationFactor:               DefaultReplicationFactor,
		DefaultRetentionMillis:                 DefaultRetentionMillis,
		DefaultEventRetryInitialIntervalMillis: DefaultEventRetryInitialIntervalMillis,
		DefaultEventRetryTimeMillisMax:         DefaultEventRetryTimeMillisMax,
		DefaultExponentialBackoff:              DefaultExponentialBackoff,
		DefaultKafkaConsumers:                  DefaultKafkaConsumers,
		DispatcherCpuLimit:                     resource.MustParse(DispatcherCpuLimit),
		DispatcherCpuRequest:                   resource.MustParse(DispatcherCpuRequest),
		DispatcherMemoryLimit:                  resource.MustParse(DispatcherMemoryLimit),
		DispatcherMemoryRequest:                resource.MustParse(DispatcherMemoryRequest),
		ChannelMemoryRequest:                   resource.MustParse(ChannelMemoryRequest),
		ChannelMemoryLimit:                     resource.MustParse(ChannelMemoryLimit),
		ChannelCpuRequest:                      resource.MustParse(ChannelCpuRequest),
		ChannelCpuLimit:                        resource.MustParse(ChannelCpuLimit),
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
		channel.Spec.Subscribable = &eventingduckv1alpha1.Subscribable{
			Subscribers: []eventingduckv1alpha1.SubscriberSpec{
				{
					UID:           SubscriberName,
					SubscriberURI: SubscriberURI,
				},
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
	channel := GetNewChannelWithProvisionedStatus(includeFinalizer, includeSpecProperties, includeSubscribers, true, true)
	channel.DeletionTimestamp = &DeletedTimestamp
	if !includeFinalizer {
		channel.ObjectMeta.Finalizers = nil
	}
	return channel
}

// Utility Function For Creating A Test Knative Messaging Channel
func GetNewKnativeMessagingChannel(channelApiVersion string, channelKind string) *messagingv1alpha1.Channel {

	// Create The Specified Knative Messaging Channel
	messagingChannel := &messagingv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: messagingv1alpha1.SchemeGroupVersion.String(),
			Kind:       constants.KnativeChannelKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NamespaceName,
			Name:      ChannelName,
		},
		Spec: messagingv1alpha1.ChannelSpec{
			ChannelTemplate: &eventingduck.ChannelTemplateSpec{
				TypeMeta: metav1.TypeMeta{
					APIVersion: channelApiVersion,
					Kind:       channelKind,
				},
			},
		},
	}

	// Return The Knative Messaging Channel
	return messagingChannel
}

// Utility Function For Creating A Test Channel With Provisioned / NotProvisioned Status
func GetNewChannelWithProvisionedStatus(includeFinalizer bool, includeSpecProperties bool, includeSubscribers bool, provisioned bool, ready bool) *kafkav1alpha1.KafkaChannel {
	channel := GetNewChannel(includeFinalizer, includeSpecProperties, includeSubscribers)
	if provisioned {
		channel.Status.InitializeConditions()
		if ready {
			channel.Status.SetAddress(&apis.URL{
				Scheme: "http",
				Host:   eventingNames.ServiceHostName(channel.Name+"-channel", channel.Namespace),
			})
			channel.Status.MarkChannelServiceTrue()
		} else {
			channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service Failed: %s", ServiceMockCreatesErrorMessage))
		}
	}
	return channel
}

// Utility Function For Creating A K8S Channel Service For The Test Channel
func GetNewK8sChannelService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ChannelDeploymentName,
			Namespace: NamespaceName,
			Labels: map[string]string{
				"channel": ChannelName,
				"k8s-app": "knative-kafka-channels",
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewChannelOwnerRef(),
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
	replicas := int32(1)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ChannelDeploymentName,
			Namespace: NamespaceName,
			Labels: map[string]string{
				"app": ChannelDeploymentName,
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewChannelOwnerRef(),
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
					Containers: []corev1.Container{
						{
							Name:  ChannelDeploymentName,
							Image: ChannelImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "server",
									ContainerPort: int32(80),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HTTP_PORT",
									Value: strconv.Itoa(constants.HttpPortNumber),
								},
								{
									Name:  "METRICS_PORT",
									Value: strconv.Itoa(MetricsPort),
								},
								{
									Name:  "KAFKA_TOPIC",
									Value: topicName,
								},
								{
									Name:  "CLIENT_ID",
									Value: topicName,
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
func GetNewSubscription(namespaceName string, subscriberName string, includeAnnotations bool, includeFinalizer bool, eventStartTime string) *eventingv1alpha1.Subscription {

	// Create The Specified Subscription
	subscription := &messagingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       constants.KnativeSubscriptionKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespaceName,
			Name:      subscriberName,
		},
		Spec: messagingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				APIVersion: kafkav1alpha1.SchemeGroupVersion.String(),
				Kind:       constants.KafkaChannelKind,
				Name:       ChannelName,
			},
			Subscriber: &eventingv1alpha1.SubscriberSpec{
				URI: &SubscriberURI,
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
func GetNewSubscriptionDeleted(namespaceName string, subscriberName string, includeAnnotations bool, includeFinalizer bool) *messagingv1alpha1.Subscription {
	subscription := GetNewSubscription(namespaceName, subscriberName, includeAnnotations, includeFinalizer, EventStartTime)
	subscription.DeletionTimestamp = &DeletedTimestamp
	return subscription
}

// Utility Function For Creating A Test Subscription With Indirect Knative Messaging Channel
func GetNewSubscriptionIndirectChannel(namespaceName string, subscriberName string, includeAnnotations bool, includeFinalizer bool, eventStartTime string, deleted bool) *messagingv1alpha1.Subscription {

	// Get The Default / Base Subscription
	var subscription *messagingv1alpha1.Subscription
	if deleted == true {
		subscription = GetNewSubscriptionDeleted(namespaceName, subscriberName, includeAnnotations, includeFinalizer)
	} else {
		subscription = GetNewSubscription(namespaceName, subscriberName, includeAnnotations, includeFinalizer, eventStartTime)
	}

	// Modify The Subscription To Use Indirect Knative Messaging Channel
	subscription.Spec.Channel.APIVersion = messagingv1alpha1.SchemeGroupVersion.String()
	subscription.Spec.Channel.Kind = constants.KnativeChannelKind

	// Return The Subscription
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
			Namespace: NamespaceName,
			Labels: map[string]string{
				"channel":    ChannelName,
				"dispatcher": "true",
				"k8s-app":    "knative-kafka-dispatchers",
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewSubscriptionOwnerRef(),
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
	replicas := int32(1)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DispatcherDeploymentName,
			Namespace: NamespaceName,
			Labels: map[string]string{
				"app":        DispatcherDeploymentName,
				"dispatcher": "true",
				"channel":    ChannelName,
			},
			OwnerReferences: []metav1.OwnerReference{
				GetNewSubscriptionOwnerRef(),
			},
			Annotations: map[string]string{
				"knativekafka.kyma-project.io/EventStartTime": EventStartTime,
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
					Containers: []corev1.Container{
						{
							Name:  DispatcherDeploymentName,
							Image: DispatcherImage,
							Env: []corev1.EnvVar{
								{
									Name:  "METRICS_PORT",
									Value: strconv.Itoa(MetricsPort),
								},
								{
									Name:  "KAFKA_GROUP_ID",
									Value: SubscriberName,
								},
								{
									Name:  "KAFKA_TOPIC",
									Value: topicName,
								},
								{
									Name:  "KAFKA_CONSUMERS",
									Value: strconv.Itoa(DefaultKafkaConsumers),
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
									Name:  "SUBSCRIBER_URI",
									Value: SubscriberURI,
								},
								{
									Name:  "EXPONENTIAL_BACKOFF",
									Value: strconv.FormatBool(DefaultExponentialBackoff),
								},
								{
									Name:  "INITIAL_RETRY_INTERVAL",
									Value: strconv.Itoa(DefaultEventRetryInitialIntervalMillis),
								},
								{
									Name:  "MAX_RETRY_TIME",
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
func GetNewChannelOwnerRef() metav1.OwnerReference {
	blockOwnerDeletion := true
	isController := true
	return metav1.OwnerReference{
		APIVersion:         "knativekafka.kyma-project.io/v1alpha1",
		Kind:               constants.KafkaChannelKind,
		Name:               ChannelName,
		UID:                "",
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

// Create A New OwnerReference Model For The Test Subscription
func GetNewSubscriptionOwnerRef() metav1.OwnerReference {
	blockOwnerDeletion := true
	isController := true
	return metav1.OwnerReference{
		APIVersion:         "eventing.knative.dev/v1alpha1",
		Kind:               "Subscription",
		Name:               SubscriberName,
		UID:                "",
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

// Create A New ChannelSubscriberSpec For Test Channel/Subscription
func GetChannelSubscriberSpec() *eventingduckv1alpha1.SubscriberSpec {
	return &eventingduckv1alpha1.SubscriberSpec{
		UID:           "SubscriberName",
		SubscriberURI: SubscriberURI,
	}
}
