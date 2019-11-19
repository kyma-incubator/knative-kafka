package k8s

import (
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // NOTE - Required When Authorizing go-client With GCP !
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path/filepath"
)

// Get A Test Client With Specified State
func GetTestKubernetesClient(initialState ...runtime.Object) kubernetes.Interface {
	return fake.NewSimpleClientset(initialState...)
}

// Initialize The Kubernetes Client (Expected To Be Called From main())
func GetKubernetesClient(logger *zap.Logger) kubernetes.Interface {

	// Get The K8S Config (In/Out Cluster)
	k8sConfig := getK8SConfig(logger)

	// Create A New Kubernetes Client From K8S Config
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		logger.Fatal("Failed To Initialize Kubernetes Client", zap.Error(err))
	}
	if k8sClient == nil {
		logger.Fatal("Created Nil Kubernetes Client")
	}

	// Update The Kubernetes Client Reference
	logger.Info("Successfully Initialized Kubernetes Client")
	return k8sClient
}

// Get The K8S Config (Cluster Internal / External)
func getK8SConfig(logger *zap.Logger) *rest.Config {

	// First Attempt In-Cluster Config (Normal Production Use Case)
	k8sConfig, err := rest.InClusterConfig()
	if err == nil && k8sConfig != nil {
		log.Logger().Info("Successfully Acquired Kubernetes In-Cluster Config")
		return k8sConfig
	} else {
		logger.Warn("Failed To Get Kubernetes In-Cluster Config", zap.Any("Config", k8sConfig), zap.Error(err))
	}

	// Get The KUBECONFIG Path
	kubeconfig := getKubeConfig()

	// Second Attempt Out-Of-Cluster Config (Development Environment Use Case)
	k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err == nil && k8sConfig != nil {
		logger.Info("Successfully Acquired Kubernetes Out-Of-Cluster Config", zap.String("$KUBECONFIG", kubeconfig))
		return k8sConfig
	} else {
		logger.Warn("Failed To Get Kubernetes Out-Of-Cluster Config", zap.Any("Config", k8sConfig), zap.Error(err), zap.String("$KUBECONFIG", kubeconfig))
	}

	// Log Fatal Failure To Get Cluster Config And Return
	logger.Fatal("Failed To Acquire Any Kubernetes Cluster Config - Terminating!")
	return nil

}

// Determine Appropriate KUBECONFIG Value (EnvVar or Default)
func getKubeConfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {

		// KUBECONFIG Environment Variable Not Specified - Default To Current Users ~/.kube/config
		currentUser, err := user.Current()
		if err != nil || currentUser == nil {
			log.Logger().Error("Environment Variable 'KUBECONFIG' Not Set - Failed To Get Current User", zap.Error(err))
		} else {
			log.Logger().Info("Environment Variable 'KUBECONFIG' Not Set - Defaulting To ~/.kube/config")
			kubeconfig = filepath.Join(currentUser.HomeDir, ".kube", "config")
		}
	}
	return kubeconfig
}
