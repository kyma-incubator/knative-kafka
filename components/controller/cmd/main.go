package main

import (
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"flag"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/prometheus"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/controller"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/controller/kafkachannel"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"go.uber.org/zap"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"knative.dev/pkg/signals"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strconv"
)

// The Main Function (Go Command)
func main() {

	// Initialize The Logger
	logger := log.Logger()
	defer safeLoggerSync(logger)
	flag.Parse()

	// Get The Environment Variables
	environment, err := env.GetEnvironment(logger)
	if err != nil {
		logger.Error("Failed To Load Environment Variables - Terminating!", zap.Error(err))
		os.Exit(1)
	}

	// Start The Prometheus Metrics Server (Prometheus)
	metricsServer := prometheus.NewMetricsServer(strconv.Itoa(environment.MetricsPort), "/metrics")
	metricsServer.Start()

	// Get The K8S Client Config (In-Cluster, etc)
	logger.Info("Determining K8S Client Config")
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error("Failed To Determine K8S Client Config - Terminating!", zap.Error(err))
		os.Exit(1)
	}

	// Create A Manager For The Controller(s)
	//    TODO - Should probably use the following kubebuilder logic to configure the controller-runtime's MetricsBindAddress.
	//         - This might require adjustments to the ServiceMonitor config in the Helm chart.
	//         - Instead we're manually starting Prometheus above and stopping below on shutdown.
	// 		var metricsAddr string
	// 		flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	// 		flag.Parse()
	// 		mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	logger.Info("Creating Controller Manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		logger.Error("Failed To Create Controller Manager", zap.Error(err))
		os.Exit(1)
	}

	// Add KafkaChannel Custom Types To The Manager's Scheme
	logger.Info("Configuring Manager's Schemes")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logger.Fatal("Failed To Add Knative-Kafka Custom Types To Manager's Scheme", zap.Error(err))
	}

	// Add Knative Messaging Custom Types To The Manager's Scheme
	err = messagingv1alpha1.AddToScheme(mgr.GetScheme())
	if err != nil {
		logger.Fatal("Failed To Add Knative Messaging Custom Types To Manager's Scheme", zap.Error(err))
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

	// Add All Controllers To Manager (Currently Just The KafkaChannel Controller)
	logger.Info("Adding Controllers To Manager")
	if err := controller.AddToManager(mgr); err != nil {
		logger.Fatal("Failed To Add Controllers To Manager", zap.Error(err))
	}

	// Setup Signals In Order To Handle Shutdown Signal Gracefully
	stopChannel := signals.SetupSignalHandler()

	// Start All Controllers & Block Until StopChannel Is Closed
	logger.Info("Starting Knative Kafka Controllers)")
	err = mgr.Start(stopChannel)
	if err != nil {
		logger.Fatal("Failed To Start Manager's Controllers!", zap.Error(err))
	}

	// Give The Channel Controller A Chance To Shutdown
	kafkachannel.Shutdown(logger)

	// Shutdown The Prometheus Metrics Server
	metricsServer.Stop()
}

// Safe (Handles Errors) Logger Sync For Deferred Usage
func safeLoggerSync(logger *zap.Logger) {
	err := logger.Sync()
	if err != nil {
		logger.Error("Failed To Sync Logger", zap.Error(err))
	}
}
