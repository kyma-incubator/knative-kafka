package kafkasecret

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
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/reconciler"
	"strconv"
)

// TODO - search everywhere for extra "channelDeploymentService" and change to "channelService"

// Reconcile The "Channel" Inbound For The Specified Kafka Secret
func (r *Reconciler) reconcileChannel(secret *corev1.Secret) error {

	// Get Secret Specific Logger
	logger := util.SecretLogger(r.Logger.Desugar(), secret)

	// Reconcile The Channel's Service
	serviceErr := r.reconcileChannelService(secret)
	if serviceErr != nil {
		r.Recorder.Eventf(secret, corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile Channel Service: %v", serviceErr)
		logger.Error("Failed To Reconcile Channel Service", zap.Error(serviceErr))
	} else {
		logger.Info("Successfully Reconciled Channel Service")
	}

	// Reconcile The Channel's Deployment
	deploymentErr := r.reconcileChannelDeployment(secret)
	if deploymentErr != nil {
		r.Recorder.Eventf(secret, corev1.EventTypeWarning, event.ChannelDeploymentReconciliationFailed.String(), "Failed To Reconcile Channel Deployment: %v", deploymentErr)
		logger.Error("Failed To Reconcile Channel Deployment", zap.Error(deploymentErr))
	} else {
		logger.Info("Successfully Reconciled Channel Deployment")
	}

	// Reconcile Channel's KafkaChannel Status
	statusErr := r.reconcileKafkaChannelStatus(secret, serviceErr == nil, deploymentErr == nil)
	if statusErr != nil {
		r.Recorder.Eventf(secret, corev1.EventTypeWarning, event.ChannelStatusReconciliationFailed.String(), "Failed To Reconcile Channel's KafkaChannel Status: %v", statusErr)
		logger.Error("Failed To Reconcile KafkaChannel Status", zap.Error(statusErr))
	} else {
		logger.Info("Successfully Reconciled KafkaChannel Status")
	}

	// Return Results
	if serviceErr != nil || deploymentErr != nil || statusErr != nil {
		return fmt.Errorf("failed to reconcile channel resources")
	} else {
		return nil // Success
	}
}

//
// Kafka Channel Service
//

// Reconcile The Kafka Channel Service
func (r *Reconciler) reconcileChannelService(secret *corev1.Secret) error {

	// Attempt To Get The Service Associated With The Specified Secret
	service, err := r.getChannelService(secret)
	if err != nil {

		// If The Service Was Not Found - Then Create A New One For The Secret
		if errors.IsNotFound(err) {

			// Then Create The New Kafka Channel Service
			r.Logger.Info("Channel Service Not Found - Creating New One")
			service = r.newChannelService(secret)
			service, err = r.KubeClientSet.CoreV1().Services(service.Namespace).Create(service)
			if err != nil {
				r.Logger.Error("Failed To Create Channel Service", zap.Error(err))
				// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
				//channel.Status.MarkChannelDeploymentServiceFailed("ChannelDeploymentServiceFailed", fmt.Sprintf("Channel Deployment Service Failed: %s", err))
				return err
			} else {
				r.Logger.Info("Successfully Created Channel Service")
				// Continue To Update Channel Status
			}

		} else {

			// Failed In Attempt To Get Kafka Channel Service From K8S
			r.Logger.Error("Failed To Get Channel Service", zap.Error(err))
			// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
			//channel.Status.MarkChannelDeploymentServiceFailed("ChannelDeploymentServiceFailed", fmt.Sprintf("Channel Deployment Service Failed: %s", err))
			return err
		}
	} else {
		// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
		r.Logger.Info("Successfully Verified Channel Service")
		// Continue To Update Channel Status
	}

	// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
	//// Update Channel Status
	//channel.Status.MarkChannelDeploymentServiceTrue()

	// Return Success
	return nil
}

// Get The Kafka Service Associated With The Specified Channel
func (r *Reconciler) getChannelService(secret *corev1.Secret) (*corev1.Service, error) {

	// Get The Dispatcher Deployment Name For The Channel - Use Same For Service
	deploymentName := util.ChannelDeploymentDnsSafeName(secret.Name)

	// Get The Service By Namespace / Name
	service := &corev1.Service{}
	service, err := r.serviceLister.Services(constants.KnativeEventingNamespace).Get(deploymentName)

	// Return The Results
	return service, err
}

// Create Kafka Service Model For The Specified Channel
func (r *Reconciler) newChannelService(secret *corev1.Secret) *corev1.Service {

	// Get The Dispatcher Deployment Name For The Channel - Use Same For Service
	deploymentName := util.ChannelDeploymentDnsSafeName(secret.Name)

	// Create & Return The Service Model
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       constants.ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				constants.KafkaChannelChannelLabel:   "true",                               // Allows for identification of KafkaChannels
				constants.K8sAppChannelSelectorLabel: constants.K8sAppChannelSelectorValue, // Prometheus ServiceMonitor (See Helm Chart)
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewSecretOwnerReference(secret),
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
					Name:       constants.MetricsPortName,
					Port:       int32(r.environment.MetricsPort),
					TargetPort: intstr.FromInt(r.environment.MetricsPort),
				},
			},
			Selector: map[string]string{
				constants.AppLabel: deploymentName, // Matches Deployment Label Key/Value
			},
		},
	}
}

//
// KafkaChannel Deployment - The Kafka Producer Implementation
//

// Reconcile The Kafka Channel Deployment
func (r *Reconciler) reconcileChannelDeployment(secret *corev1.Secret) error {

	// Attempt To Get The KafkaChannel Deployment Associated With The Specified Channel
	deployment, err := r.getChannelDeployment(secret)
	if err != nil {

		// If The KafkaChannel Deployment Was Not Found - Then Create A New Deployment For The Channel
		if errors.IsNotFound(err) {

			// Then Create The New Deployment
			r.Logger.Info("KafkaChannel Deployment Not Found - Creating New One")
			deployment, err = r.newChannelDeployment(secret)
			if err != nil {
				r.Logger.Error("Failed To Create KafkaChannel Deployment YAML", zap.Error(err))
				// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
				//channel.Status.MarkChannelDeploymentFailed("ChannelDeploymentFailed", fmt.Sprintf("Channel Deployment Failed: %s", err))
				return err
			} else {
				deployment, err = r.KubeClientSet.AppsV1().Deployments(deployment.Namespace).Create(deployment)
				if err != nil {
					r.Logger.Error("Failed To Create KafkaChannel Deployment", zap.Error(err))
					// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
					//channel.Status.MarkChannelDeploymentFailed("ChannelDeploymentFailed", fmt.Sprintf("Channel Deployment Failed: %s", err))
					return err
				} else {
					r.Logger.Info("Successfully Created KafkaChannel Deployment")
					// Continue To Update Channel Status
				}
			}

		} else {

			// Failed In Attempt To Get Deployment From K8S
			r.Logger.Error("Failed To Get KafkaChannel Deployment", zap.Error(err))
			// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
			//channel.Status.MarkChannelDeploymentFailed("ChannelDeploymentFailed", fmt.Sprintf("Channel Deployment Failed: %s", err))
			return err
		}
	} else {
		// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
		r.Logger.Info("Successfully Verified Channel Deployment")
		// Continue To Update Channel Status
	}

	// TODO - no status on secret - but... will what to do with the status on the KafkaChannel ?
	//// Update Channel Status
	//channel.Status.MarkChannelDeploymentTrue()

	// Return Success
	return nil
}

// Get The Kafka Channel Deployment Associated With The Specified Secret
func (r *Reconciler) getChannelDeployment(secret *corev1.Secret) (*appsv1.Deployment, error) {

	// Get The Channel Deployment Name (One Channel Deployment Per Kafka Auth Secret)
	deploymentName := util.ChannelDeploymentDnsSafeName(secret.Name)

	// Get The Channel Deployment By Namespace / Name
	deployment := &appsv1.Deployment{}
	deployment, err := r.deploymentLister.Deployments(constants.KnativeEventingNamespace).Get(deploymentName)

	// Return The Results
	return deployment, err
}

// Create Kafka Channel Deployment Model For The Specified Secret
func (r *Reconciler) newChannelDeployment(secret *corev1.Secret) (*appsv1.Deployment, error) {

	// Get The Channel Deployment Name (One Channel Deployment Per Kafka Auth Secret)
	deploymentName := util.ChannelDeploymentDnsSafeName(secret.Name)

	// Replicas Int Value For De-Referencing
	replicas := int32(r.environment.ChannelReplicas)

	// Create The Channel Container Environment Variables
	channelEnvVars, err := r.channelDeploymentEnvVars(secret)
	if err != nil {
		r.Logger.Error("Failed To Create Channel Deployment Environment Variables", zap.Error(err))
		return nil, err
	}

	// Create & Return The Channel's Deployment
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       constants.DeploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: constants.KnativeEventingNamespace,
			Labels: map[string]string{
				constants.AppLabel:                 deploymentName, // Matches Service Selector Key/Value Below
				constants.KafkaChannelChannelLabel: "true",         // Allows for identification of KafkaChannels
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewSecretOwnerReference(secret),
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
								InitialDelaySeconds: constants.ChannelLivenessDelay,
								PeriodSeconds:       constants.ChannelLivenessPeriod,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Port: intstr.FromInt(constants.HealthPort),
										Path: health.ReadinessPath,
									},
								},
								InitialDelaySeconds: constants.ChannelReadinessDelay,
								PeriodSeconds:       constants.ChannelReadinessPeriod,
							},
							Image: r.environment.ChannelImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "server",
									ContainerPort: int32(constants.HttpContainerPortNumber),
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

// Create The Kafka Channel Deployment's Env Vars
func (r *Reconciler) channelDeploymentEnvVars(secret *corev1.Secret) ([]corev1.EnvVar, error) {

	// Create The Channel Deployment EnvVars
	envVars := []corev1.EnvVar{
		{
			Name:  env.MetricsPortEnvVarKey,
			Value: strconv.Itoa(r.environment.MetricsPort),
		},
		{
			Name:  env.HealthPortEnvVarKey,
			Value: strconv.Itoa(constants.HealthPort),
		},
	}

	// Append The Kafka Brokers As Env Var
	envVars = append(envVars, corev1.EnvVar{
		Name: env.KafkaBrokerEnvVarKey,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				Key:                  constants.KafkaSecretDataKeyBrokers,
			},
		},
	})

	// Append The Kafka Username As Env Var
	envVars = append(envVars, corev1.EnvVar{
		Name: env.KafkaUsernameEnvVarKey,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				Key:                  constants.KafkaSecretDataKeyUsername,
			},
		},
	})

	// Append The Kafka Password As Env Var
	envVars = append(envVars, corev1.EnvVar{
		Name: env.KafkaPasswordEnvVarKey,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				Key:                  constants.KafkaSecretDataKeyPassword,
			},
		},
	})

	// Return The Channel Deployment EnvVars Array
	return envVars, nil
}

//
// KafkaChannel Status Reconciliation
//

// Reconcile KafkaChannel Status With Specified Channel Service/Deployment State
func (r *Reconciler) reconcileKafkaChannelStatus(secret *corev1.Secret, validService bool, validDeployment bool) error {

	// Get A Secret Logger (With The Valid Service/Deployment State
	logger := util.SecretLogger(r.Logger.Desugar(), secret).With(zap.Bool("Service", validService), zap.Bool("Deployment", validDeployment))

	// Create Selector With Requirement For KafkaSecret Labels With Value Of Specified Secret Name
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(constants.KafkaSecretLabel, selection.Equals, []string{secret.Name})
	if err != nil {
		logger.Error("Failed To Create Selector Requirement For Kafka Secret Label", zap.Error(err)) // Should Never Happen
		return err
	}
	selector.Add(*requirement)

	// List The KafkaChannels Which Match The Selector (All Namespaces)
	kafkaChannels, err := r.kafkachannelLister.List(selector)
	if err != nil {
		logger.Error("Failed To List KafkaChannels For Kafka Secret", zap.Error(err))
		return err
	}

	// Update All The KafkaChannels Status As Specified (Process All Regardless Of Error)
	statusUpdateErrors := false
	for _, kafkaChannel := range kafkaChannels {
		if kafkaChannel != nil {
			err := r.updateKafkaChannelStatus(kafkaChannel, validService, validDeployment)
			if err != nil {
				logger.Error("Failed To Update KafkaChannel Status", zap.Error(err))
				statusUpdateErrors = true
			}
		}
	}

	// Return Status Update Error
	if statusUpdateErrors {
		return fmt.Errorf("failed to update Status of one or more KafkaChannels")
	} else {
		return nil
	}
}

// Update A Single KafkaChannel's Status To Reflect The Specified Channel Service/Deployment State
func (r *Reconciler) updateKafkaChannelStatus(originalChannel *knativekafkav1alpha1.KafkaChannel, validService bool, validDeployment bool) error {

	// Get A KafkaChannel Logger
	logger := util.ChannelLogger(r.Logger.Desugar(), originalChannel)

	// Update The KafkaChannel (Retry On Conflict - KafkaChannel Controller Will Also Be Updating KafkaChannel Status)
	return reconciler.RetryUpdateConflicts(func(attempts int) error {

		var err error

		// After First Attempt - Reload The Original KafkaChannel From K8S
		if attempts > 0 {
			originalChannel, err = r.kafkachannelLister.KafkaChannels(originalChannel.Namespace).Get(originalChannel.Name)
			if err != nil {
				logger.Error("Failed To Reload KafkaChannel For Status Update", zap.Error(err))
				return err
			}
		}

		// Clone The KafkaChannel So As Not To Perturb Informers Copy
		updatedChannel := originalChannel.DeepCopy()

		// Update Service Status Based On Specified State
		if validService {
			updatedChannel.Status.MarkChannelDeploymentServiceTrue()
		} else {
			// TODO - change these reasons ???
			updatedChannel.Status.MarkChannelDeploymentServiceFailed("ChannelDeploymentServiceFailed", fmt.Sprint("Channel Service Failed"))
		}

		// Update Deployment Status Based On Specified State
		if validDeployment {
			updatedChannel.Status.MarkChannelDeploymentTrue()
		} else {
			updatedChannel.Status.MarkChannelDeploymentFailed("ChannelDeploymentFailed", fmt.Sprint("Channel Deployment Failed"))
		}

		// If The KafkaChannel Status Changed
		if !equality.Semantic.DeepEqual(originalChannel.Status, updatedChannel.Status) {

			// Then Attempt To Update The KafkaChannel Status
			_, err = r.kafkaChannelClient.KnativekafkaV1alpha1().KafkaChannels(updatedChannel.Namespace).UpdateStatus(updatedChannel)
			if err != nil {
				logger.Error("Failed To Update KafkaChannel Status", zap.Error(err))
				return err
			} else {
				logger.Info("Successfully Updated KafkaChannel Status")
				return nil
			}

		} else {

			// Otherwise No Change To Status - Return Success
			logger.Info("Successfully Verified KafkaChannel Status")
			return nil
		}
	})

}
