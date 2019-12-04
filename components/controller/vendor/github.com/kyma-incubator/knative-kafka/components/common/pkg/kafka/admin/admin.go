package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

// Confluent Client Doesn't Code To Interfaces Or Provide Mocks So We're Wrapping Our Usage Of The AdminClient For Testing
// Also Introduced Additional Functionality To Get The Kafka Secret For A Topic
type AdminClientInterface interface {
	CreateTopics(context.Context, []kafka.TopicSpecification, ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error)
	DeleteTopics(context.Context, []string, ...kafka.DeleteTopicsAdminOption) ([]kafka.TopicResult, error)
	Close()
	GetKafkaSecretName(topicName string) string
}

// AdminClient Type Enumeration
type AdminClientType int

const (
	Kafka AdminClientType = iota
	EventHub
)

//
// Create A New Kafka AdminClient Of Specified Type - Using Credentials From Kafka Secret(s) In Specified K8S Namespace
//
// The K8S Namespace parameter indicates the Kubernetes Namespace in which the Kafka Credentials secret(s)
// will be found.  The secret(s) must contain the constants.KafkaSecretLabel label indicating it is a "Kafka Secret".
//
// For the normal Kafka use case (Confluent, etc.) there should be only one Secret with the following content...
//
//      data:
//		  brokers: SASL_SSL://<host>.<region>.aws.confluent.cloud:9092
//        username: <username>
//        password: <password>
//
// For the Azure EventHub use case there will be multiple Secrets (one per Azure Namespace) each with the following content...
//
//      data:
//        brokers: <azure-namespace>.servicebus.windows.net:9093
//        username: $ConnectionString
//        password: Endpoint=sb://<azure-namespace>.servicebus.windows.net/;SharedAccessKeyName=<shared-access-key-name>;SharedAccessKey=<shared-access-key-value>
//		  namespace: <azure-namespace>
//
// * If no authorization is required (local dev instance) then specify username and password as the empty string ""
//
func CreateAdminClient(logger *zap.Logger, adminClientType AdminClientType, k8sNamespace string) (AdminClientInterface, error) {

	// Validate Parameters
	if len(k8sNamespace) <= 0 {
		return nil, errors.New(fmt.Sprintf("required parameters not provided: k8sNamespace='%s'", k8sNamespace))
	}

	// Get A New Kafka AdminClient From ConfigMap & Return Results
	return NewAdminClientWrapper(logger, adminClientType, k8sNamespace)
}

// Kafka Function Reference Variable To Facilitate Mocking In Unit Tests
var NewAdminClientWrapper = func(logger *zap.Logger, adminClientType AdminClientType, k8sNamespace string) (AdminClientInterface, error) {

	// Create The Appropriate Type Of AdminClient
	if adminClientType == Kafka {
		return NewKafkaAdminClient(logger, k8sNamespace)
	} else if adminClientType == EventHub {
		return NewEventHubAdminClient(logger, k8sNamespace)
	} else {
		return nil, errors.New(fmt.Sprintf("received unsupported AdminClientType of %d", adminClientType))
	}
}
