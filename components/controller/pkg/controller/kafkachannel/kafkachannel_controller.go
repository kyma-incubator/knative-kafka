package kafkachannel

import (
	istiov1alpha3 "github.com/knative/pkg/apis/istio/v1alpha3"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/reconciler/kafkachannel"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Package Level Kafka AdminClient Reference (For Shutdown() Usage)
var adminClient kafkaadmin.AdminClientInterface

// Creates & Initialize A KafkaChannel Controller & Add It To Specified Manager With Default RBAC
// (Manager will set fields on the Controller and start it when the Manager is started.)
func Add(mgr manager.Manager, adminClient kafkaadmin.AdminClientInterface) error {

	// Create A Child Logger With The Controller Name
	logger := log.Logger().With(zap.String("Controller", constants.KafkaChannelControllerAgentName))

	// Get The Controller's Environment
	environment, err := env.GetEnvironment(logger)
	if err != nil {
		logger.Error("Failed To Get Controller Environment", zap.Error(err))
		return err
	}

	// Get A K8S Event Recorder From The Manager
	recorder := mgr.GetRecorder(constants.KafkaChannelControllerAgentName)

	// Create A New KafkaChannel Reconciler
	kafkaChannelReconciler := kafkachannel.NewReconciler(recorder, logger, adminClient, environment)

	// Create A KafkaChannel Controller
	c, err := controller.New(constants.KafkaChannelControllerAgentName, mgr, controller.Options{Reconciler: kafkaChannelReconciler})
	if err != nil {
		logger.Error("Failed To Create Channel Controller!", zap.Error(err))
		return err
	}

	// Configure Controller To Watch Channel CustomResources
	err = c.Watch(&source.Kind{Type: &kafkav1alpha1.KafkaChannel{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of Channel CustomResources", zap.Error(err), zap.Any("type", &kafkav1alpha1.KafkaChannel{}))
		return err
	}

	// Configure Controller To Watch K8s Services That Are Owned By Channels
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{OwnerType: &kafkav1alpha1.KafkaChannel{}, IsController: true})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of K8S Services Owned By Channels", zap.Error(err))
		return err
	}

	// Configure Controller To Watch Istio VirtualServices That Are Owned By Channels
	err = c.Watch(&source.Kind{Type: &istiov1alpha3.VirtualService{}}, &handler.EnqueueRequestForOwner{OwnerType: &kafkav1alpha1.KafkaChannel{}, IsController: true})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of Istio VirtualServices Owned By Channels", zap.Error(err))
		return err
	}

	// Configure Controller To Watch K8s Deployments That Are Owned By Channels
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{OwnerType: &kafkav1alpha1.KafkaChannel{}, IsController: true})
	if err != nil {
		logger.Error("Failed To Configure Controller Watch Of K8S Deployments Owned By Channels", zap.Error(err))
		return err
	}

	// Return Success
	return nil
}

// Graceful Shutdown Hook
func Shutdown(logger *zap.Logger) {
	logger.Info("Controller Shutting Down", zap.String("Controller", constants.KafkaChannelControllerAgentName))
	adminClient.Close()
}
