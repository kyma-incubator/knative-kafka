package kafkachannel

import (
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/health"
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
	"knative.dev/pkg/logging"
	"knative.dev/pkg/system"
	"strconv"
)

//
// Reconcile The Dispatcher (Kafka Consumer) For The Specified KafkaChannel
//
func (r *Reconciler) reconcileDispatcher(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Get Channel Specific Logger
	logger := util.ChannelLogger(r.Logger.Desugar(), channel)

	// Reconcile The Dispatcher's Service (For Prometheus Only)
	serviceErr := r.reconcileDispatcherService(channel)
	if serviceErr != nil {
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.DispatcherServiceReconciliationFailed.String(), "Failed To Reconcile Dispatcher Service: %v", serviceErr)
		logger.Error("Failed To Reconcile Dispatcher Service", zap.Error(serviceErr))
	} else {
		logger.Info("Successfully Reconciled Dispatcher Service")
	}

	// Reconcile The Dispatcher's Deployment
	deploymentErr := r.reconcileDispatcherDeployment(channel)
	if deploymentErr != nil {
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile Dispatcher Deployment: %v", deploymentErr)
		logger.Error("Failed To Reconcile Dispatcher Deployment", zap.Error(deploymentErr))
		channel.Status.MarkDispatcherDeploymentFailed("DispatcherDeploymentFailed", fmt.Sprintf("Dispatcher Deployment Failed: %s", deploymentErr))
	} else {
		logger.Info("Successfully Reconciled Dispatcher Deployment")
		channel.Status.MarkDispatcherDeploymentTrue()
	}

	// Return Results
	if serviceErr != nil || deploymentErr != nil {
		return fmt.Errorf("failed to reconcile dispatcher resources")
	} else {
		return nil
	}
}

//
// Dispatcher Service
//

// Reconcile The Dispatcher Service
func (r *Reconciler) reconcileDispatcherService(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Attempt To Get The Dispatcher Service Associated With The Specified Channel
	service, err := r.getDispatcherService(channel)
	if err != nil {

		// If The Service Was Not Found - Then Create A New One For The Channel
		if errors.IsNotFound(err) {
			r.Logger.Info("Dispatcher Service Not Found - Creating New One")
			service = r.newDispatcherService(channel)
			service, err = r.KubeClientSet.CoreV1().Services(service.Namespace).Create(service)
			if err != nil {
				r.Logger.Error("Failed To Create Dispatcher Service", zap.Error(err))
				return err
			} else {
				r.Logger.Info("Successfully Created Dispatcher Service")
				return nil
			}
		} else {
			r.Logger.Error("Failed To Get Dispatcher Service", zap.Error(err))
			return err
		}
	} else {
		r.Logger.Info("Successfully Verified Dispatcher Service")
		return nil
	}
}

// Get The Dispatcher Service Associated With The Specified Channel
func (r *Reconciler) getDispatcherService(channel *knativekafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Get The Dispatcher Service Name
	serviceName := util.DispatcherDnsSafeName(channel)

	// Get The Service By Namespace / Name
	service := &corev1.Service{}
	service, err := r.serviceLister.Services(constants.KnativeEventingNamespace).Get(serviceName)

	// Return The Results
	return service, err
}

// Create Dispatcher Service Model For The Specified Subscription
func (r *Reconciler) newDispatcherService(channel *knativekafkav1alpha1.KafkaChannel) *corev1.Service {

	// Get The Dispatcher Service Name For The Channel
	serviceName := util.DispatcherDnsSafeName(channel)

	// Create & Return The Service Model
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       constants.ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				constants.KafkaChannelDispatcherLabel:   "true",                                  // Identifies the Service as being a KafkaChannel "Dispatcher"
				constants.KafkaChannelNameLabel:         channel.Name,                            // Identifies the Service's Owning KafkaChannel's Name
				constants.KafkaChannelNamespaceLabel:    channel.Namespace,                       // Identifies the Service's Owning KafkaChannel's Namespace
				constants.K8sAppDispatcherSelectorLabel: constants.K8sAppDispatcherSelectorValue, // Prometheus ServiceMonitor (See Helm Chart)
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelOwnerReference(channel),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       constants.MetricsPortName,
					Port:       int32(r.environment.MetricsPort),
					TargetPort: intstr.FromInt(r.environment.MetricsPort),
				},
			},
			Selector: map[string]string{
				constants.AppLabel: serviceName, // Matches Deployment Label Key/Value
			},
		},
	}
}

//
// Dispatcher Deployment
//

// Reconcile The Dispatcher Deployment
func (r *Reconciler) reconcileDispatcherDeployment(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Attempt To Get The Dispatcher Deployment Associated With The Specified Channel
	deployment, err := r.getDispatcherDeployment(channel)
	if err != nil {

		// If The Dispatcher Deployment Was Not Found - Then Create A New Deployment For The Channel
		if errors.IsNotFound(err) {

			// Then Create The New Deployment
			r.Logger.Info("Dispatcher Deployment Not Found - Creating New One")
			deployment, err = r.newDispatcherDeployment(channel)
			if err != nil {
				r.Logger.Error("Failed To Create Dispatcher Deployment YAML", zap.Error(err))
				return err
			} else {
				deployment, err = r.KubeClientSet.AppsV1().Deployments(deployment.Namespace).Create(deployment)
				if err != nil {
					r.Logger.Error("Failed To Create Dispatcher Deployment", zap.Error(err))
					return err
				} else {
					r.Logger.Info("Successfully Created Dispatcher Deployment")
					return nil
				}
			}
		} else {
			// Failed In Attempt To Get Deployment From K8S
			r.Logger.Error("Failed To Get KafkaChannel Deployment", zap.Error(err))
			return err
		}
	} else {
		r.Logger.Info("Successfully Verified Dispatcher Deployment")
		return nil
	}
}

// Get The Dispatcher Deployment Associated With The Specified Channel
func (r *Reconciler) getDispatcherDeployment(channel *knativekafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

	// Get The Dispatcher Deployment Name For The Channel
	deploymentName := util.DispatcherDnsSafeName(channel)

	// Get The Dispatcher Deployment By Namespace / Name
	deployment := &appsv1.Deployment{}
	deployment, err := r.deploymentLister.Deployments(constants.KnativeEventingNamespace).Get(deploymentName)

	// Return The Results
	return deployment, err
}

// Create Dispatcher Deployment Model For The Specified Channel
func (r *Reconciler) newDispatcherDeployment(channel *knativekafkav1alpha1.KafkaChannel) (*appsv1.Deployment, error) {

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

	// Create The Dispatcher's Deployment
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       constants.DeploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				constants.AppLabel:                    deploymentName,    // Matches K8S Service Selector Key/Value Below
				constants.KafkaChannelDispatcherLabel: "true",            // Identifies the Deployment as being a KafkaChannel "Dispatcher"
				constants.KafkaChannelNameLabel:       channel.Name,      // Identifies the Deployment's Owning KafkaChannel's Name
				constants.KafkaChannelNamespaceLabel:  channel.Namespace, // Identifies the Deployment's Owning KafkaChannel's Namespace
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelOwnerReference(channel),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					constants.AppLabel: deploymentName, // Matches Template ObjectMeta Pods
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						constants.AppLabel: deploymentName, // Matched By Deployment Selector Above
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: r.environment.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name: deploymentName,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Port: intstr.FromInt(constants.HealthPort),
										Path: health.LivenessPath,
									},
								},
								InitialDelaySeconds: constants.DispatcherLivenessDelay,
								PeriodSeconds:       constants.DispatcherLivenessPeriod,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Port: intstr.FromInt(constants.HealthPort),
										Path: health.ReadinessPath,
									},
								},
								InitialDelaySeconds: constants.DispatcherReadinessDelay,
								PeriodSeconds:       constants.DispatcherReadinessPeriod,
							},
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

	// Return The Dispatcher's Deployment
	return deployment, nil
}

// Create The Dispatcher Container's Env Vars
func (r *Reconciler) dispatcherDeploymentEnvVars(channel *knativekafkav1alpha1.KafkaChannel) ([]corev1.EnvVar, error) {

	// Get The TopicName For Specified Channel
	topicName := util.TopicName(channel)

	// Create The Dispatcher Deployment EnvVars
	envVars := []corev1.EnvVar{
		{
			Name:  system.NamespaceEnvKey,
			Value: constants.KnativeEventingNamespace,
		},
		{
			Name:  logging.ConfigMapNameEnv,
			Value: logging.ConfigMapName(),
		},
		{
			Name:  env.MetricsPortEnvVarKey,
			Value: strconv.Itoa(r.environment.MetricsPort),
		},
		{
			Name:  env.HealthPortEnvVarKey,
			Value: strconv.Itoa(constants.HealthPort),
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
