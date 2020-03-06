package kafkachannel

import (
	"fmt"
	kafkautil "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	knativekafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingNames "knative.dev/eventing/pkg/reconciler/names"
	eventingUtils "knative.dev/eventing/pkg/utils"
	"knative.dev/pkg/apis"
)

/* TODO - status !!!!

- lastTransitionTime: "2020-03-05T17:26:53Z"
  status: Unknown
  type: ChannelDeploymentReady

- lastTransitionTime: "2020-03-05T17:26:53Z"
  status: Unknown
  type: ChannelDeploymentServiceReady
*/

// Reconcile The "Channel" Inbound For The Specified Channel
func (r *Reconciler) reconcileChannel(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Get Channel Specific Logger
	logger := util.ChannelLogger(r.Logger.Desugar(), channel)

	// Reconcile The KafkaChannel's Service
	channelServiceErr := r.reconcileKafkaChannelService(channel)
	if channelServiceErr != nil {
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelServiceReconciliationFailed.String(), "Failed To Reconcile Channel Service (Channel): %v", channelServiceErr)
		logger.Error("Failed To Reconcile KafkaChannel Service", zap.Error(channelServiceErr))
	} else {
		logger.Info("Successfully Reconciled KafkaChannel Service")
	}

	// Return Results
	if channelServiceErr != nil {
		return fmt.Errorf("failed to reconcile channel components")
	} else {
		return nil // Success
	}
}

//
// KafkaChannel Kafka Channel Service
//
// One K8S Service per KafkaChannel, in the same namespace as the KafkaChannel, with an
// ExternalName reference to the single K8S Service in the knative-eventing namespace
// for the Channel Deployment/Pods.
//

// Reconcile The KafkaChannel Service
func (r *Reconciler) reconcileKafkaChannelService(channel *knativekafkav1alpha1.KafkaChannel) error {

	// Attempt To Get The Service Associated With The Specified Channel
	service, err := r.getKafkaChannelService(channel)
	if err != nil {

		// If The Service Was Not Found - Then Create A New One For The Channel
		if errors.IsNotFound(err) {
			r.Logger.Info("KafkaChannel Service Not Found - Creating New One")
			service = r.newKafkaChannelService(channel)
			service, err = r.KubeClientSet.CoreV1().Services(service.Namespace).Create(service)
			if err != nil {
				r.Logger.Error("Failed To Create KafkaChannel Service", zap.Error(err))
				channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service Failed: %s", err))
				return err
			} else {
				r.Logger.Info("Successfully Created KafkaChannel Service")
				// Continue To Update Channel Status
			}
		} else {
			r.Logger.Error("Failed To Get KafkaChannel Service", zap.Error(err))
			channel.Status.MarkChannelServiceFailed("ChannelServiceFailed", fmt.Sprintf("Channel Service Failed: %s", err))
			return err
		}
	} else {
		r.Logger.Info("Successfully Verified KafkaChannel Service")
		// Continue To Update Channel Status
	}

	// Update Channel Status
	channel.Status.MarkChannelServiceTrue()
	channel.Status.SetAddress(&apis.URL{
		Scheme: "http",
		Host:   eventingNames.ServiceHostName(service.Name, service.Namespace),
	})

	// Return Success
	return nil
}

// Get The KafkaChannel Service Associated With The Specified Channel
func (r *Reconciler) getKafkaChannelService(channel *knativekafkav1alpha1.KafkaChannel) (*corev1.Service, error) {

	// Get The KafkaChannel Service Name
	serviceName := kafkautil.AppendKafkaChannelServiceNameSuffix(channel.Name)

	// Get The Service By Namespace / Name
	service := &corev1.Service{}
	service, err := r.serviceLister.Services(channel.Namespace).Get(serviceName)

	// Return The Results
	return service, err
}

// Create KafkaChannel Service Model For The Specified Channel
func (r *Reconciler) newKafkaChannelService(channel *knativekafkav1alpha1.KafkaChannel) *corev1.Service {

	// Get The KafkaChannel Service Name
	serviceName := kafkautil.AppendKafkaChannelServiceNameSuffix(channel.Name)

	// Get The Dispatcher Service Name For The Channel (One Channel Service Per KafkaChannel Instance)
	deploymentName := util.ChannelDeploymentDnsSafeName(r.kafkaSecretName(channel))
	deploymentServiceAddress := fmt.Sprintf("%s.%s.svc.%s", deploymentName, constants.KnativeEventingNamespace, eventingUtils.GetClusterDomainName())

	// Create & Return The Service Model
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       constants.ServiceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,       // Must Match KafkaChannel For HOST Parsing In Channel Implementation!
			Namespace: channel.Namespace, // Must Match KafkaChannel For HOST Parsing In Channel Implementation!
			Labels: map[string]string{
				constants.KafkaChannelLabel:          channel.Name,
				constants.KafkaChannelChannelLabel:   "true",                               // Allows for identification of KafkaChannels
				constants.K8sAppChannelSelectorLabel: constants.K8sAppChannelSelectorValue, // Prometheus ServiceMonitor (See Helm Chart)
			},
			OwnerReferences: []metav1.OwnerReference{
				util.NewChannelOwnerReference(channel),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: deploymentServiceAddress,
		},
	}
}

//
// Utility Functions (Uses AdminClient)
//

// Get The Kafka Auth Secret Corresponding To The Specified KafkaChannel
func (r *Reconciler) kafkaSecretName(channel *knativekafkav1alpha1.KafkaChannel) string {
	return r.adminClient.GetKafkaSecretName(util.TopicName(channel))
}
