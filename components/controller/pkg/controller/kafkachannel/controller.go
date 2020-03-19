package kafkachannel

import (
	"context"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkachannelv1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	knativekafkaclientsetinjection "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/client"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/informers/knativekafka/v1alpha1/kafkachannel"
	kafkachannelreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/kafkasecretinformer"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/eventing/pkg/logging"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"os"
)

// Package Level Kafka AdminClient Reference (For Shutdown() Usage)
var adminClient kafkaadmin.AdminClientInterface

// Create A New KafkaChannel Controller
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {

	// Get A Logger
	logger := logging.FromContext(ctx)

	// Get The Needed Informers
	kafkachannelInformer := kafkachannel.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)
	kafkaSecretInformer := kafkasecretinformer.Get(ctx)

	// Load The Environment Variables
	environment, err := env.GetEnvironment(logger)
	if err != nil {
		logger.Panic("Failed To Load Environment Variables - Terminating!", zap.Error(err))
		os.Exit(1)
	}

	// Determine The Kafka AdminClient Type (Assume Kafka Unless Azure EventHubs Are Specified)
	kafkaAdminClientType := kafkaadmin.Kafka
	if environment.KafkaProvider == env.KafkaProviderValueAzure {
		kafkaAdminClientType = kafkaadmin.EventHub
	}

	// Get The Kafka AdminClient
	adminClient, err := kafkaadmin.CreateAdminClient(logger, kafkaAdminClientType, constants.KnativeEventingNamespace)
	if adminClient == nil || err != nil {
		logger.Fatal("Failed To Create Kafka AdminClient", zap.Error(err))
	}

	// Create A KafkaChannel Reconciler
	r := &Reconciler{
		Base:                  reconciler.NewBase(ctx, constants.KafkaChannelControllerAgentName, cmw),
		environment:           environment,
		knativekafkaClientSet: knativekafkaclientsetinjection.Get(ctx),
		kafkachannelLister:    kafkachannelInformer.Lister(),
		kafkachannelInformer:  kafkachannelInformer.Informer(),
		deploymentLister:      deploymentInformer.Lister(),
		serviceLister:         serviceInformer.Lister(),
		adminClient:           adminClient,
	}

	// Create A New KafkaChannel Controller Impl With The Reconciler
	controllerImpl := kafkachannelreconciler.NewImpl(ctx, r)

	//
	// Configure The Informers' EventHandlers
	//
	// Note - The use of EnqueueLabelOfNamespaceScopedResource() is to facilitate cross-namespace OwnerReference
	//        management and relies upon the reconciler creating the Services/Deployments with appropriate labels.
	//        Kubernetes OwnerReferences are not intended to be cross-namespace and thus don't include the namespace
	//        information.
	//
	r.Logger.Info("Setting Up EventHandlers")
	kafkachannelInformer.Informer().AddEventHandler(
		controller.HandleAll(controllerImpl.Enqueue),
	)
	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kafkachannelv1alpha1.SchemeGroupVersion.WithKind(constants.KafkaChannelKind)),
		Handler:    controller.HandleAll(controllerImpl.EnqueueLabelOfNamespaceScopedResource(constants.KafkaChannelNamespaceLabel, constants.KafkaChannelNameLabel)),
	})
	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kafkachannelv1alpha1.SchemeGroupVersion.WithKind(constants.KafkaChannelKind)),
		Handler:    controller.HandleAll(controllerImpl.EnqueueLabelOfNamespaceScopedResource(constants.KafkaChannelNamespaceLabel, constants.KafkaChannelNameLabel)),
	})
	kafkaSecretInformer.Informer().AddEventHandler(
		controller.HandleAll(r.resetKafkaAdminClient(kafkaAdminClientType)),
	)

	// Return The KafkaChannel Controller Impl
	return controllerImpl
}

// Graceful Shutdown Hook
func Shutdown() {
	if adminClient != nil {
		adminClient.Close()
	}
}

// Recreate The Kafka AdminClient On The Reconciler (Useful To Reload Cache Which Is Not Yet Exposed)
func (r *Reconciler) resetKafkaAdminClient(kafkaAdminClientType kafkaadmin.AdminClientType) func(obj interface{}) {
	return func(obj interface{}) {
		adminClient, err := kafkaadmin.CreateAdminClient(r.Logger.Desugar(), kafkaAdminClientType, constants.KnativeEventingNamespace)
		if adminClient == nil || err != nil {
			r.Logger.Error("Failed To Re-Create Kafka AdminClient", zap.Error(err))
		} else {
			r.Logger.Info("Successfully Re-Created Kafka AdminClient")
			r.adminClient = adminClient
		}
	}
}
