package k8s

import (
	"context"
	injectionclient "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/injection/client"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/system"
	"log"
)

// TODO - i think we should roll the initialization of the kafkachannel client into this context and make a general InitializeContext() fn

//
// Initialize The Specified Context With A K8S Client & Logger (ConfigMap Watcher)
//
// Note - This logic represents a stepping stone on our path towards alignment with the Knative eventing-contrib implementations.
//        We are not an "injected controller" in the knative-eventing injection framework, but still want to leverage that
//        implementation as much as possible to ease future refactoring.  This will allow us to use the default knative-eventing
//        logging configuration and dynamic updating.  To that end, we are setting up a basic context ourselves that mirrors
//        what the injection framework would have created.
//
func LoggingContext(ctx context.Context, component string, masterUrl string, kubeconfigPath string) context.Context {

	// Create The K8S Configuration (In-Cluster By Default / Cmd Line Flags For Out-Of-Cluster Usage)
	k8sConfig, err := k8sclientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		log.Fatalf("Failed To Build Kubernetes Config: %v", err)
	}

	// Create A New Kubernetes Client From The K8S Configuration
	k8sClient := kubernetes.NewForConfigOrDie(k8sConfig)

	// Put The Kubernetes Client Into The Context Where The Injection Framework Expects It
	ctx = context.WithValue(ctx, injectionclient.Key{}, k8sClient)

	// Get The Logging Config From Knative SharedMain
	loggingConfig, err := sharedmain.GetLoggingConfig(ctx)
	if err != nil {
		log.Fatalf("Failed To Read/Parse Logging Configuration: %v", err)
	}

	// Create A New Logger From The Logging Config & Add To Context
	logger, atomicLevel := logging.NewLoggerFromConfig(loggingConfig, component)
	ctx = logging.WithLogger(ctx, logger)

	// Create A Watcher On The Logging ConfigMap & Dynamically Update Log Levels
	cmw := configmap.NewInformedWatcher(k8sClient, system.Namespace()) // Note - Have removed cmLabelReqs filtering here.
	if _, err := k8sClient.CoreV1().ConfigMaps(system.Namespace()).Get(logging.ConfigMapName(), metav1.GetOptions{}); err == nil {
		logger.Info("Setting Logging ConfigMap Watcher")
		cmw.Watch(logging.ConfigMapName(), logging.UpdateLevelFromConfigMap(logger, atomicLevel, component))
	} else if !apierrors.IsNotFound(err) {
		logger.Fatalf("Error Reading Logging ConfigMap %q", logging.ConfigMapName(), zap.Error(err))
	}

	// Start The Logging ConfigMap Watcher
	if err := cmw.Start(ctx.Done()); err != nil {
		logger.Fatalw("Failed To Start ConfigMap Watcher", zap.Error(err))
	}

	// Return The Initialized Context
	return ctx
}
