package kafkachannel

import (
	"context"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/client/injection/kube/informers/core/v1/service"

	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkachannelv1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/informers/knativekafka/v1alpha1/kafkachannel"
	kafkachannelreconciler "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/reconciler/knativekafka/v1alpha1/kafkachannel"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/logging"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"os"
)

// Package Level Kafka AdminClient Reference (For Shutdown() Usage)
var adminClient kafkaadmin.AdminClientInterface

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {

	// Get A Logger
	logger := logging.FromContext(ctx)

	// Get The Needed Informers
	kafkachannelInformer := kafkachannel.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)

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
		Base: reconciler.NewBase(ctx, constants.KafkaChannelControllerAgentName, cmw),

		environment:          environment,
		kafkachannelLister:   kafkachannelInformer.Lister(),
		kafkachannelInformer: kafkachannelInformer.Informer(),
		deploymentLister:     deploymentInformer.Lister(),
		serviceLister:        serviceInformer.Lister(),
		adminClient:          adminClient,
	}

	// Create A New KafkaChannel Controller Impl With The Reconciler
	controllerImpl := kafkachannelreconciler.NewImpl(ctx, r)

	// Configure The Informers' EventHandlers
	r.Logger.Info("Setting Up EventHandlers")
	kafkachannelInformer.Informer().AddEventHandler(controller.HandleAll(controllerImpl.Enqueue))
	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kafkachannelv1alpha1.SchemeGroupVersion.WithKind("KafkaChannel")),
		Handler:    controller.HandleAll(controllerImpl.EnqueueControllerOf),
	})
	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kafkachannelv1alpha1.SchemeGroupVersion.WithKind("KafkaChannel")),
		Handler:    controller.HandleAll(controllerImpl.EnqueueControllerOf),
	})

	// Return The KafkaChannel Controller Impl
	return controllerImpl
}

// Graceful Shutdown Hook
func Shutdown() {
	adminClient.Close()
}
