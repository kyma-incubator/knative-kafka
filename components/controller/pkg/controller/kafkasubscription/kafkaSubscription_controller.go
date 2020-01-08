package kafkasubscription

import (
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/reconciler/kafkasubscription"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Creates & Initialize A "KafkaSubscription" (Knative Eventing Subscription CRD) Controller & Add It To Specified Manager With Default RBAC
// (Manager will set fields on the Controller and start it when the Manager is started.)
func Add(mgr manager.Manager, adminClient kafkaadmin.AdminClientInterface) error {

	// Create A Child Logger With The Controller Name
	logger := log.Logger().With(zap.String("Controller", constants.KafkaSubscriptionControllerAgentName))

	// Get The Controller's Environment
	environment, err := env.GetEnvironment(logger)
	if err != nil {
		logger.Error("Failed To Get Controller Environment", zap.Error(err))
		return err
	}

	// Get A K8S Event Recorder From The Manager
	recorder := mgr.GetEventRecorderFor(constants.KafkaSubscriptionControllerAgentName)

	// Create A Kafka (Knative) Subscription Reconciler
	kafkaSubscriptionReconciler := kafkasubscription.NewReconciler(recorder, logger, adminClient, environment)

	// Create A Subscription Controller
	c, err := controller.New(constants.KafkaSubscriptionControllerAgentName, mgr, controller.Options{Reconciler: kafkaSubscriptionReconciler})
	if err != nil {
		logger.Error("Failed To Create Kafka Subscription Controller!", zap.Error(err))
		return err
	}

	// Configure Controller To Watch Subscription CustomResources
	err = c.Watch(&source.Kind{Type: &messagingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of Subscription CustomResources", zap.Error(err), zap.Any("type", &messagingv1alpha1.Subscription{}))
		return err
	}

	// Configure Controller To Watch K8s Services That Are Owned By Subscriptions
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{OwnerType: &messagingv1alpha1.Subscription{}, IsController: true})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of K8S Services Owned By Subscriptions", zap.Error(err))
		return err
	}

	// Configure Controller To Watch K8s Deployments That Are Owned By Subscriptions
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{OwnerType: &messagingv1alpha1.Subscription{}, IsController: true})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of K8S Deployments Owned By Subscriptions", zap.Error(err))
		return err
	}

	// Return Success
	return nil
}

// Graceful Shutdown Hook
func Shutdown(logger *zap.Logger) {
	logger.Info("Controller Shutting Down", zap.String("Controller", constants.KafkaSubscriptionControllerAgentName))
}
