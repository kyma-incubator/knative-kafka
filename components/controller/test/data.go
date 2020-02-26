package test

import (
	"fmt"
	kafkautil "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	reconcilertesting "knative.dev/pkg/reconciler/testing"
	"strconv"
	"time"
)

// Constants
const (
	// Prometheus MetricsPort
	MetricsPortName = "metrics"

	// Environment Test Data
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

	// Channel Test Data
	KafkaChannelNamespace = "kafkachannel-namespace"
	KafkaChannelName      = "kafkachannel-name"
	KafkaChannelKey       = KafkaChannelNamespace + "/" + KafkaChannelName
	ChannelDeploymentName = KafkaSecret + "-channel"
	TopicName             = KafkaChannelNamespace + "." + KafkaChannelName

	// ChannelSpec Test Data
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

// KafkaChannelOption Enables Customization Of A KafkaChannel
type KafkaChannelOption func(*knativekafkav1alpha1.KafkaChannel)

// Utility Function For Creating A Custom KafkaChannel For Testing
func NewKnativeKafkaChannel(options ...KafkaChannelOption) *knativekafkav1alpha1.KafkaChannel {

	// Create The Specified KafkaChannel
	kafkachannel := &knativekafkav1alpha1.KafkaChannel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: knativekafkav1alpha1.SchemeGroupVersion.String(),
			Kind:       constants.KafkaChannelKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: KafkaChannelNamespace,
			Name:      KafkaChannelName,
		},
		Spec: knativekafkav1alpha1.KafkaChannelSpec{
			NumPartitions:     NumPartitions,
			ReplicationFactor: ReplicationFactor,
			RetentionMillis:   RetentionMillis,
		},
	}

	// Apply The Specified KafkaChannel Customizations
	for _, option := range options {
		option(kafkachannel)
	}

	// Return The Test KafkaChannel
	return kafkachannel
}

// Set The KafkaChannel's Status To Initialized State
func WithInitKafkaChannelConditions(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.InitializeConditions()
}

// Set The KafkaChannel's DeletionTimestamp To Current Time
func WithKafkaChannelDeleted(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	deleteTime := metav1.NewTime(time.Unix(1e9, 0))
	kafkachannel.ObjectMeta.SetDeletionTimestamp(&deleteTime)
}

// Set The KafkaChannel's Finalizer
func WithFinalizer(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.ObjectMeta.Finalizers = []string{"kafkachannels.knativekafka.kyma-project.io"}
}

// Set The KafkaChannel's Address
func WithKafkaChannelAddress(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.SetAddress(&apis.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s-kafkachannel.%s.svc.cluster.local", KafkaChannelName, KafkaChannelNamespace),
	})
}

// Set The KafkaChannel's Channel Service As READY
func WithKafkaChannelChannelServiceReady(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkChannelServiceTrue()
}

// Set The KafkaChannel's Channel Services As Failed
func WithKafkaChannelChannelServiceFailed(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service Failed: inducing failure for create services"))
}

// Set The KafkaChannel's Deployment Service As READY
func WithKafkaChannelDeploymentServiceReady(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkChannelDeploymentServiceTrue()
}

// Set The KafkaChannel's Deployment Service As Failed
func WithKafkaChannelDeploymentServiceFailed(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkChannelDeploymentServiceFailed("ChannelDeploymentServiceFailed", "Channel Deployment Service Failed: inducing failure for create services")
}

// Set The KafkaChannel's Channel Deployment As READY
func WithKafkaChannelChannelDeploymentReady(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkChannelDeploymentTrue()
}

// Set The KafkaChannel's Channel Deployment As Failed
func WithKafkaChannelChannelDeploymentFailed(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkChannelDeploymentFailed("ChannelDeploymentFailed", "Channel Deployment Failed: inducing failure for create deployments")
}

// Set The KafkaChannel's Dispatcher Deployment As READY
func WithKafkaChannelDispatcherDeploymentReady(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkDispatcherDeploymentTrue()
}

// Set The KafkaChannel's Dispatcher Deployment As Failed
func WithKafkaChannelDispatcherDeploymentFailed(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkDispatcherDeploymentFailed("DispatcherDeploymentFailed", "Dispatcher Deployment Failed: inducing failure for create deployments")
}

// Set The KafkaChannel's Topic READY
func WithKafkaChannelTopicReady(kafkachannel *knativekafkav1alpha1.KafkaChannel) {
	kafkachannel.Status.MarkTopicTrue()
}

// Utility Function For Creating A Custom KafkaChannel "Channel" Service For Testing
func NewKafkaChannelChannelService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       constants.ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkautil.AppendKafkaChannelServiceNameSuffix(KafkaChannelName),
			Namespace: KafkaChannelNamespace,
			Labels: map[string]string{
				"kafkachannel":         KafkaChannelName,
				"kafkachannel-channel": "true",
				"k8s-app":              "knative-kafka-channels",
			},
			OwnerReferences: []metav1.OwnerReference{
				NewChannelOwnerRef(true),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: ChannelDeploymentName + "." + constants.KnativeEventingNamespace + ".svc.cluster.local",
		},
	}
}

// Utility Function For Creating A Custom KafkaChannel "Deployment" Service For Testing
func NewKafkaChannelDeploymentService() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       constants.ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ChannelDeploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"k8s-app":              "knative-kafka-channels",
				"kafkachannel-channel": "true",
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
func NewKafkaChannelChannelDeployment() *appsv1.Deployment {
	replicas := int32(ChannelReplicas)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       constants.DeploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ChannelDeploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"app":                  ChannelDeploymentName,
				"kafkachannel-channel": "true",
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

// Utility Function For Creating A Custom KafkaChannel Dispatcher Service For Testing
func NewKafkaChannelDispatcherService() *corev1.Service {

	// Get The Expected Service Name For The Test KafkaChannel
	serviceName := util.DispatcherDnsSafeName(&knativekafkav1alpha1.KafkaChannel{
		ObjectMeta: metav1.ObjectMeta{Namespace: KafkaChannelNamespace, Name: KafkaChannelName},
	})

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       constants.ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"kafkachannel":            KafkaChannelName,
				"kafkachannel-dispatcher": "true",
				"k8s-app":                 "knative-kafka-dispatchers",
			},
			OwnerReferences: []metav1.OwnerReference{
				NewChannelOwnerRef(true),
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
				"app": serviceName,
			},
		},
	}
}

// Utility Function For Creating A Custom KafkaChannel Dispatcher Deployment For Testing
func NewKafkaChannelDispatcherDeployment() *appsv1.Deployment {

	// Get The Expected Dispatcher & Topic Names For The Test KafkaChannel
	sparseKafkaChannel := &knativekafkav1alpha1.KafkaChannel{ObjectMeta: metav1.ObjectMeta{Namespace: KafkaChannelNamespace, Name: KafkaChannelName}}
	dispatcherName := util.DispatcherDnsSafeName(sparseKafkaChannel)
	topicName := util.TopicName(sparseKafkaChannel)

	// Replicas Int Reference
	replicas := int32(DispatcherReplicas)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       constants.DeploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dispatcherName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				"app":                     dispatcherName,
				"kafkachannel-dispatcher": "true",
				"kafkachannel":            KafkaChannelName,
			},
			OwnerReferences: []metav1.OwnerReference{
				NewChannelOwnerRef(true),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dispatcherName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": dispatcherName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:  dispatcherName,
							Image: DispatcherImage,
							Env: []corev1.EnvVar{
								{
									Name:  env.MetricsPortEnvVarKey,
									Value: strconv.Itoa(MetricsPort),
								},
								{
									Name:  env.ChannelKeyEnvVarKey,
									Value: fmt.Sprintf("%s/%s", KafkaChannelNamespace, KafkaChannelName),
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

// Utility Function For Creating A New OwnerReference Model For The Test Channel
func NewChannelOwnerRef(isController bool) metav1.OwnerReference {
	blockOwnerDeletion := true
	return metav1.OwnerReference{
		APIVersion:         "knativekafka.kyma-project.io/v1alpha1",
		Kind:               constants.KafkaChannelKind,
		Name:               KafkaChannelName,
		UID:                "",
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

// Utility Function For Creating A PatchActionImpl For The Finalizer Patch Command
func NewFinalizerPatchActionImpl() clientgotesting.PatchActionImpl {
	return clientgotesting.PatchActionImpl{
		ActionImpl: clientgotesting.ActionImpl{
			Namespace:   KafkaChannelNamespace,
			Verb:        "patch",
			Resource:    schema.GroupVersionResource{Group: knativekafkav1alpha1.SchemeGroupVersion.Group, Version: knativekafkav1alpha1.SchemeGroupVersion.Version, Resource: "kafkachannels"},
			Subresource: "",
		},
		Name:      KafkaChannelName,
		PatchType: "application/merge-patch+json",
		Patch:     []byte(`{"metadata":{"finalizers":["kafkachannels.knativekafka.kyma-project.io"],"resourceVersion":""}}`),
		// Above finalizer name matches package private "defaultFinalizerName" constant in injection/reconciler/knativekafka/v1alpha1/kafkachannel ;)
	}
}

// Utility Function For Creating A Successful KafkaChannel Reconciled Event
func NewKafkaChannelSuccessfulReconciliationEvent() string {
	return reconcilertesting.Eventf(corev1.EventTypeNormal, event.KafkaChannelReconciled.String(), `KafkaChannel Reconciled Successfully: "%s/%s"`, KafkaChannelNamespace, KafkaChannelName)
}

// Utility Function For Creating A Failed KafkaChannel Reconciled Event
func NewKafkaChannelFailedReconciliationEvent() string {
	return reconcilertesting.Eventf(corev1.EventTypeWarning, "InternalError", constants.ReconciliationFailedError)
}
