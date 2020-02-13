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

func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	kafkachannelInformer := kafkachannel.Get(ctx)
	deploymentInformer := deployment.Get(ctx)
	serviceInformer := service.Get(ctx)

	logger := logging.FromContext(ctx)
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

	// Get The Kafka AdminClient (Doing At This Level So That All Controllers Can Share The Reference)
	adminClient, err := kafkaadmin.CreateAdminClient(logger, kafkaAdminClientType, constants.KnativeEventingNamespace)
	if adminClient == nil || err != nil {
		logger.Fatal("Failed To Create Kafka AdminClient", zap.Error(err))
	}

	r := &Reconciler{
		Base: reconciler.NewBase(ctx, constants.KafkaChannelControllerAgentName, cmw),

		environment:          environment,
		kafkachannelLister:   kafkachannelInformer.Lister(),
		kafkachannelInformer: kafkachannelInformer.Informer(),
		deploymentLister:     deploymentInformer.Lister(),
		serviceLister:        serviceInformer.Lister(),
		adminClient:          adminClient,
	}

	impl := kafkachannelreconciler.NewImpl(ctx, r)

	r.Logger.Info("Setting up event handlers")
	kafkachannelInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kafkachannelv1alpha1.SchemeGroupVersion.WithKind("KafkaChannel")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(kafkachannelv1alpha1.SchemeGroupVersion.WithKind("KafkaChannel")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return &controller.Impl{}
}

// Graceful Shutdown Hook
func Shutdown() {
	adminClient.Close()
}
