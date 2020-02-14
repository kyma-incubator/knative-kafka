package eventhubcache

import (
	"context"
	"fmt"
	eventhub "github.com/Azure/azure-event-hubs-go"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/k8s"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/constants"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	"testing"
)

// Test The Cache's NewCache() Constructor
func TestNewCache(t *testing.T) {

	// Test Data
	k8sNamespace := "TestK8SNamespace"

	// Test Logger
	logger := log.TestLogger()

	// Replace The GetKubernetesClient Wrapper To Provide Mock Implementation & Defer Reset
	getKubernetesClientWrapperPlaceholder := GetKubernetesClientWrapper
	GetKubernetesClientWrapper = func(logger *zap.Logger) kubernetes.Interface {
		return k8s.GetTestKubernetesClient()
	}
	defer func() { GetKubernetesClientWrapper = getKubernetesClientWrapperPlaceholder }()

	// Perform The Test
	cache := NewCache(logger, k8sNamespace)

	// Verify The Results
	assert.NotNil(t, cache)
}

// Test The Cache's Update() Functionality
func TestUpdate(t *testing.T) {

	// Test Data
	k8sNamespace1 := "TestK8SNamespace1"
	k8sNamespace2 := "TestK8SNamespace2"

	kafkaSecretName1 := "TestKafkaSecretName1"
	kafkaSecretName2 := "TestKafkaSecretName2"
	kafkaSecretName3 := "TestKafkaSecretName3"

	kafkaSecretBrokers1 := "TestKafkaSecretBrokers1"
	kafkaSecretBrokers2 := "TestKafkaSecretBrokers2"
	kafkaSecretBrokers3 := "TestKafkaSecretBrokers3"

	kafkaSecretUsername1 := "TestKafkaSecretUsername1"
	kafkaSecretUsername2 := "TestKafkaSecretUsername2"
	kafkaSecretUsername3 := "TestKafkaSecretUsername3"

	kafkaSecretPassword1 := "TestKafkaSecretPassword1"
	kafkaSecretPassword2 := "TestKafkaSecretPassword2"
	kafkaSecretPassword3 := "TestKafkaSecretPassword3"

	kafkaSecretNamespace1 := "TestKafkaSecretNamespace1"
	kafkaSecretNamespace2 := "TestKafkaSecretNamespace2"
	kafkaSecretNamespace3 := "TestKafkaSecretNamespace3"

	kafkaSecret1 := createKafkaSecret(kafkaSecretName1, k8sNamespace1, kafkaSecretBrokers1, kafkaSecretUsername1, kafkaSecretPassword1, kafkaSecretNamespace1)
	kafkaSecret2 := createKafkaSecret(kafkaSecretName2, k8sNamespace1, kafkaSecretBrokers2, kafkaSecretUsername2, kafkaSecretPassword2, kafkaSecretNamespace2)
	kafkaSecret3 := createKafkaSecret(kafkaSecretName3, k8sNamespace2, kafkaSecretBrokers3, kafkaSecretUsername3, kafkaSecretPassword3, kafkaSecretNamespace3)

	HubEntityName1 := "TestHubEntityName1"
	HubEntityName2 := "TestHubEntityName2"
	HubEntityName3 := "TestHubEntityName3"

	hubEntity1 := createEventHubEntity(HubEntityName1)
	hubEntity2 := createEventHubEntity(HubEntityName2)
	hubEntity3 := createEventHubEntity(HubEntityName3)

	// Create A Test Logger
	logger := log.TestLogger()

	// Create A Cache To Test
	cache := &Cache{
		logger:       logger,
		k8sClient:    k8s.GetTestKubernetesClient(kafkaSecret1, kafkaSecret2, kafkaSecret3),
		k8sNamespace: k8sNamespace1,
		namespaceMap: make(map[string]*Namespace),
		eventhubMap:  make(map[string]*Namespace),
	}

	// Create Some Mock HubManagers To Return EventHubs For Azure Namespace List Queries
	mockHubManager1 := &MockHubManager{ListHubEntities: []*eventhub.HubEntity{hubEntity1, hubEntity2}}
	mockHubManager2 := &MockHubManager{ListHubEntities: []*eventhub.HubEntity{hubEntity3}}

	// Replace The NewHubManagerFromConnectionString Wrapper To Provide Mock Implementation & Defer Reset
	newHubManagerFromConnectionStringWrapperPlaceholder := NewHubManagerFromConnectionStringWrapper
	NewHubManagerFromConnectionStringWrapper = func(connectionString string) (managerInterface HubManagerInterface, e error) {
		if strings.Contains(connectionString, kafkaSecretPassword1) {
			return mockHubManager1, nil
		} else if strings.Contains(connectionString, kafkaSecretPassword2) {
			return mockHubManager2, nil
		} else {
			return nil, fmt.Errorf("unexpected test connectionString '%s'", connectionString)
		}
	}
	defer func() { NewHubManagerFromConnectionStringWrapper = newHubManagerFromConnectionStringWrapperPlaceholder }()

	// Perform The Test
	err := cache.Update(context.TODO())

	// Verify Results
	assert.Nil(t, err)
	assert.Len(t, cache.eventhubMap, 3) // The Number Of HubEntities Returned From List
	cacheEventHubNamespace1 := cache.eventhubMap[hubEntity1.Name]
	cacheEventHubNamespace2 := cache.eventhubMap[hubEntity2.Name]
	cacheEventHubNamespace3 := cache.eventhubMap[hubEntity3.Name]
	assert.NotNil(t, cacheEventHubNamespace1)
	assert.NotNil(t, cacheEventHubNamespace2)
	assert.NotNil(t, cacheEventHubNamespace3)
	assert.Equal(t, kafkaSecretNamespace1, cacheEventHubNamespace1.Name)
	assert.Equal(t, kafkaSecretNamespace1, cacheEventHubNamespace2.Name)
	assert.Equal(t, kafkaSecretNamespace2, cacheEventHubNamespace3.Name)
}

// Test The Cache's AddEventHub() Functionality
func TestAddEventHub(t *testing.T) {

	// Test Data
	namespaceName1 := "TestNamespaceName1"
	namespaceName2 := "TestNamespaceName2"

	// Create A Mock HubManager
	mockHubManager := &MockHubManager{}

	// Replace The NewHubManagerFromConnectionString Wrapper To Provide Mock Implementation & Defer Reset
	newHubManagerFromConnectionStringWrapperPlaceholder := NewHubManagerFromConnectionStringWrapper
	NewHubManagerFromConnectionStringWrapper = func(connectionString string) (managerInterface HubManagerInterface, e error) {
		return mockHubManager, nil
	}
	defer func() { NewHubManagerFromConnectionStringWrapper = newHubManagerFromConnectionStringWrapperPlaceholder }()

	// Create Test Namespaces
	namespace1 := createTestNamespaceWithCount(namespaceName1, 1)

	// Create The Cache's EventHub Map
	eventHubMap := make(map[string]*Namespace)
	eventHubMap[namespaceName1] = namespace1

	// Create A Cache To Test
	cache := &Cache{
		logger:      log.TestLogger(),
		eventhubMap: eventHubMap,
	}

	// Perform The Test
	cache.AddEventHub(context.TODO(), namespaceName2, createTestNamespaceWithCount(namespaceName2, 0))

	// Verify The Results
	assert.Len(t, cache.eventhubMap, 2)
	assert.NotNil(t, cache.eventhubMap[namespaceName1])
	assert.NotNil(t, cache.eventhubMap[namespaceName2])
	assert.Equal(t, namespaceName1, cache.eventhubMap[namespaceName1].Name)
	assert.Equal(t, namespaceName2, cache.eventhubMap[namespaceName2].Name)
	assert.Equal(t, 1, cache.eventhubMap[namespaceName1].Count)
	assert.Equal(t, 1, cache.eventhubMap[namespaceName2].Count)
}

// Test The Cache's RemoveEventHub() Functionality
func TestRemoveEventHub(t *testing.T) {
	// Test Data
	namespaceName1 := "TestNamespaceName1"
	namespaceName2 := "TestNamespaceName2"

	// Create A Mock HubManager
	mockHubManager := &MockHubManager{}

	// Replace The NewHubManagerFromConnectionString Wrapper To Provide Mock Implementation & Defer Reset
	newHubManagerFromConnectionStringWrapperPlaceholder := NewHubManagerFromConnectionStringWrapper
	NewHubManagerFromConnectionStringWrapper = func(connectionString string) (managerInterface HubManagerInterface, e error) {
		return mockHubManager, nil
	}
	defer func() { NewHubManagerFromConnectionStringWrapper = newHubManagerFromConnectionStringWrapperPlaceholder }()

	// Create Test Namespaces
	namespace1 := createTestNamespaceWithCount(namespaceName1, 1)
	namespace2 := createTestNamespaceWithCount(namespaceName2, 1)

	// Create The Cache's EventHub Map
	eventhubMap := make(map[string]*Namespace)
	eventhubMap[namespaceName1] = namespace1
	eventhubMap[namespaceName2] = namespace2

	// Create A Cache To Test
	cache := &Cache{
		logger:      log.TestLogger(),
		eventhubMap: eventhubMap,
	}

	// Perform The Test
	cache.RemoveEventHub(context.TODO(), namespaceName2)

	// Verify The Results
	assert.Len(t, cache.eventhubMap, 1)
	assert.Nil(t, cache.eventhubMap[namespaceName2])
	assert.Equal(t, namespaceName1, cache.eventhubMap[namespaceName1].Name)
	assert.Nil(t, cache.eventhubMap[namespaceName2])
	assert.Equal(t, 1, namespace1.Count)
	assert.Equal(t, 0, namespace2.Count)
}

// Test The Cache's GetNamespace() Functionality
func TestGetNamespace(t *testing.T) {

	// Test Data
	namespaceName1 := "TestNamespaceName1"
	namespaceName2 := "TestNamespaceName2"
	namespaceSecret1 := "TestNamespaceSecret1"
	namespaceSecret2 := "TestNamespaceSecret2"

	// Create A Mock HubManager
	mockHubManager := &MockHubManager{}

	// Replace The NewHubManagerFromConnectionString Wrapper To Provide Mock Implementation & Defer Reset
	newHubManagerFromConnectionStringWrapperPlaceholder := NewHubManagerFromConnectionStringWrapper
	NewHubManagerFromConnectionStringWrapper = func(connectionString string) (managerInterface HubManagerInterface, e error) {
		return mockHubManager, nil
	}
	defer func() { NewHubManagerFromConnectionStringWrapper = newHubManagerFromConnectionStringWrapperPlaceholder }()

	// Create The Cache's EventHub Map
	eventHubMap := make(map[string]*Namespace)
	eventHubMap[namespaceName1] = createTestNamespaceWithSecret(namespaceName1, namespaceSecret1)
	eventHubMap[namespaceName2] = createTestNamespaceWithSecret(namespaceName2, namespaceSecret2)

	// Create A Cache To Test
	cache := &Cache{
		logger:      log.TestLogger(),
		eventhubMap: eventHubMap,
	}

	// Perform The Test
	namespace := cache.GetNamespace(namespaceName2)

	// Verify The Results
	assert.NotNil(t, namespace)
	assert.Equal(t, namespaceName2, namespace.Name)
	assert.Equal(t, namespaceSecret2, namespace.Secret)
}

// Test The Cache's GetNamespaceWithMaxCapacity() Functionality
func TestGetNamespaceWithMaxCapacity(t *testing.T) {

	// Test Data
	namespaceName1 := "TestNamespaceName1"
	namespaceName2 := "TestNamespaceName2"
	namespaceName3 := "TestNamespaceName3"
	namespaceName4 := "TestNamespaceName4"
	namespaceName5 := "TestNamespaceName5"
	namespaceCount1 := 4
	namespaceCount2 := 2
	namespaceCount3 := 0
	namespaceCount4 := 2
	namespaceCount5 := 1

	// Create A Mock HubManager
	mockHubManager := &MockHubManager{}

	// Replace The NewHubManagerFromConnectionString Wrapper To Provide Mock Implementation & Defer Reset
	newHubManagerFromConnectionStringWrapperPlaceholder := NewHubManagerFromConnectionStringWrapper
	NewHubManagerFromConnectionStringWrapper = func(connectionString string) (managerInterface HubManagerInterface, e error) {
		return mockHubManager, nil
	}
	defer func() { NewHubManagerFromConnectionStringWrapper = newHubManagerFromConnectionStringWrapperPlaceholder }()

	// Create The Cache's Namespace Map
	namespaceMap := make(map[string]*Namespace)
	namespaceMap[namespaceName1] = createTestNamespaceWithCount(namespaceName1, namespaceCount1)
	namespaceMap[namespaceName2] = createTestNamespaceWithCount(namespaceName2, namespaceCount2)
	namespaceMap[namespaceName3] = createTestNamespaceWithCount(namespaceName3, namespaceCount3)
	namespaceMap[namespaceName4] = createTestNamespaceWithCount(namespaceName4, namespaceCount4)
	namespaceMap[namespaceName5] = createTestNamespaceWithCount(namespaceName5, namespaceCount5)

	// Create A Cache To Test
	cache := &Cache{
		logger:       log.TestLogger(),
		namespaceMap: namespaceMap,
	}

	// Perform The Test
	namespace := cache.GetNamespaceWithMaxCapacity()

	// Verify Results
	assert.NotNil(t, namespace)
	assert.Equal(t, namespaceName3, namespace.Name)
	assert.Equal(t, namespaceCount3, namespace.Count)
}

//
// Utilities
//

// Create K8S Kafka Secret With Specified Config
func createKafkaSecret(name string, namespace string, brokers string, username string, password string, eventHubNamespace string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				constants.KafkaSecretLabel: "true",
			},
		},
		Data: map[string][]byte{
			constants.KafkaSecretKeyBrokers:   []byte(brokers),
			constants.KafkaSecretKeyUsername:  []byte(username),
			constants.KafkaSecretKeyPassword:  []byte(password),
			constants.KafkaSecretKeyNamespace: []byte(eventHubNamespace),
		},
	}
}

// Create EventHub HubEntity with Specified Name
func createEventHubEntity(name string) *eventhub.HubEntity {
	return &eventhub.HubEntity{Name: name}
}

// Create Cache Namespace With Name & Secret Only For Convenience
func createTestNamespaceWithSecret(name string, secret string) *Namespace {
	return NewNamespace(name, name, name, secret, 0)
}

// Create Cache Namespace With Name & Count Only For Convenience
func createTestNamespaceWithCount(name string, count int) *Namespace {
	return NewNamespace(name, name, name, name, count)
}
