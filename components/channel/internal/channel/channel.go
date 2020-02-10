package channel

import (
	"errors"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	knativekafkaclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	knativekafkainformers "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/informers/externalversions"
	knativekafkalisters "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"
	eventingChannel "knative.dev/eventing/pkg/channel"
	knativecontroller "knative.dev/pkg/controller"
)

// Package Variables
var (
	kafkaChannelLister knativekafkalisters.KafkaChannelLister
	stopChan           chan struct{}
)

// Wrapper Around KnativeKafka Client Creation To Facilitate Unit Testing
var getKnativeKafkaClient = func(masterUrl string, kubeconfigPath string) (knativekafkaclientset.Interface, error) {

	// Create The K8S Configuration (In-Cluster With Cmd Line Flags For Out-Of-Cluster Usage)
	k8sConfig, err := k8sclientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		log.Logger().Error("Failed To Build Kubernetes Config", zap.Error(err))
		return nil, err
	}

	// Create A New KnativeKafka Client From The K8S Config & Return The Result
	return knativekafkaclientset.NewForConfigOrDie(k8sConfig), nil
}

// Initialize The KafkaChannel Lister Singleton
func InitializeKafkaChannelLister(masterUrl string, kubeconfigPath string) error {

	// Create The K8S KnativeKafka Client For KafkaChannels
	client, err := getKnativeKafkaClient(masterUrl, kubeconfigPath)
	if err != nil {
		log.Logger().Error("Failed To Create KnativeKafka Client", zap.Error(err))
		return err
	}

	// Create A New KafkaChannel SharedInformerFactory For ALL Namespaces (Default Resync Is 10 Hrs)
	sharedInformerFactory := knativekafkainformers.NewSharedInformerFactory(client, knativecontroller.DefaultResyncPeriod)

	// Initialize The Stop Channel (Close If Previously Created)
	Close()
	stopChan = make(chan struct{})

	// Get A KafkaChannel Informer From The SharedInformerFactory - Start The Informer & Wait For It
	kafkaChannelInformer := sharedInformerFactory.Knativekafka().V1alpha1().KafkaChannels()
	go kafkaChannelInformer.Informer().Run(stopChan)
	sharedInformerFactory.WaitForCacheSync(stopChan)

	// Get A KafkaChannel Lister From The Informer
	kafkaChannelLister = kafkaChannelInformer.Lister()

	// Return Success
	log.Logger().Info("Successfully Initialized KafkaChannel Lister")
	return nil
}

// Validate The Specified ChannelReference Is For A Valid (Existing / READY) KafkaChannel
func ValidateKafkaChannel(channelReference eventingChannel.ChannelReference) error {

	// Update The Logger With ChannelReference
	logger := log.Logger().With(zap.Any("ChannelReference", channelReference))

	// Validate The Specified Channel Reference
	if len(channelReference.Name) <= 0 || len(channelReference.Namespace) <= 0 {
		logger.Warn("Invalid KafkaChannel - Invalid ChannelReference")
		return errors.New("invalid ChannelReference specified")
	}

	// Attempt To Get The KafkaChannel From The KafkaChannel Lister
	kafkaChannel, err := kafkaChannelLister.KafkaChannels(channelReference.Namespace).Get(channelReference.Name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Note - Returning Knative UnknownChannelError Type For EventReceiver.StartHTTP()
			//        Ideally we'd be able to populate the UnknownChannelError's ChannelReference
			//        but once again Knative has made this private.
			logger.Warn("Invalid KafkaChannel - Not Found")
			return &eventingChannel.UnknownChannelError{}
		} else {
			logger.Error("Invalid KafkaChannel - Failed To Find", zap.Error(err))
			return err
		}
	}

	// Check KafkaChannel READY Status
	if !kafkaChannel.Status.IsReady() {
		logger.Info("Invalid KafkaChannel - Not READY")
		return errors.New("channel status not READY")
	}

	// Return Valid KafkaChannel
	logger.Debug("Valid KafkaChannel - Found & READY")
	return nil
}

// Close The Channel Lister (Stop Processing)
func Close() {
	if stopChan != nil {
		log.Logger().Info("Closing Informer's Stop Channel")
		close(stopChan)
	}
}
