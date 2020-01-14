package controller

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/informers/externalversions/knativekafka/v1alpha1"
	listers "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/listers/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/dispatcher/internal/dispatcher"
	"go.uber.org/zap"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/logging"

	"knative.dev/pkg/controller"
)

const (
	// ReconcilerName is the name of the reconciler.
	ReconcilerName = "KafkaChannels"
)

// Reconciler reconciles Kafka Channels.
type Reconciler struct {
	dispatcher           *dispatcher.Dispatcher
	Logger               *zap.Logger
	kafkachannelInformer cache.SharedIndexInformer
	kafkachannelLister   listers.KafkaChannelLister
	impl                 *controller.Impl
}

var _ controller.Reconciler = Reconciler{}

// NewController initializes the controller and is called by the generated code.
// Registers event handlers to enqueue events.
func NewController(logger *zap.Logger, dispatcher *dispatcher.Dispatcher, kafkachannelInformer v1alpha1.KafkaChannelInformer) *controller.Impl {

	r := &Reconciler{
		Logger:               logger,
		dispatcher:           dispatcher,
		kafkachannelInformer: kafkachannelInformer.Informer(),
		kafkachannelLister:   kafkachannelInformer.Lister(),
	}
	r.impl = controller.NewImpl(r, r.Logger.Sugar(), ReconcilerName)

	r.Logger.Info("Setting Up Event Handlers")

	// Watch for kafka channels.
	kafkachannelInformer.Informer().AddEventHandler(controller.HandleAll(r.impl.Enqueue))

	return r.impl
}

func (r Reconciler) Reconcile(ctx context.Context, key string) error {

	r.Logger.Info("Reconcile", zap.String("key", key))

	// Convert the namespace/name string into a distinct namespace and name.
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logging.FromContext(ctx).Error("invalid resource key")
		return nil
	}

	// Get the KafkaChannel resource with this namespace/name.
	original, err := r.kafkachannelLister.KafkaChannels(namespace).Get(name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			r.Logger.Warn("KafkaChannel No Longer Exists", zap.String("namespace", namespace), zap.String("name", name))
			return nil
		}
		r.Logger.Error("Error Retrieving KafkaChannel", zap.Error(err), zap.String("namespace", namespace), zap.String("name", name))
		return err
	}

	// Only Reconcile KafkaChannel Associated With This Dispatcher
	if r.dispatcher.ChannelName != name {
		return nil
	}

	if original.Spec.Subscribable != nil {
		subscriptions := make([]dispatcher.Subscription, 0)
		for _, subscriber := range original.Spec.Subscribable.Subscribers {
			groupId := fmt.Sprintf("kafka.%s", subscriber.UID)
			subscriptions = append(subscriptions, dispatcher.Subscription{URI: subscriber.SubscriberURI, GroupId: groupId})
			r.Logger.Info("Adding Subscriber, Consumer Group", zap.String("groupId", groupId), zap.String("URI", subscriber.SubscriberURI))
		}

		r.dispatcher.UpdateSubscriptions(subscriptions)
	}

	return nil
}
