package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/k8s"
	util2 "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin/util"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/constants"
	"github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//
// This is an implementation of the AdminClient interface backed by the Confluent API. This is largely a pass-through
// to the Confluent client with some additional functionality layered on top.  Primarily this layer exists
//

// Ensure The KafkaAdminClient Struct Implements The AdminClientInterface
var _ AdminClientInterface = &KafkaAdminClient{}

// Kafka AdminClient Definition
type KafkaAdminClient struct {
	logger      *zap.Logger
	namespace   string
	kafkaSecret string
	adminClient ConfluentAdminClientInterface
}

// GetKubernetesClient Wrapper To Facilitate Unit Testing
var GetKubernetesClientWrapper = func(logger *zap.Logger) kubernetes.Interface {
	return k8s.GetKubernetesClient(logger)
}

// Confluent AdminClient Interface - Adding Our Own Wrapping Interface To Facilitate Testing
type ConfluentAdminClientInterface interface {
	CreateTopics(ctx context.Context, topicSpecifications []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error)
	DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) ([]kafka.TopicResult, error)
	Close()
}

// Verify The Confluent AdminClient Implements Our Interface
var _ ConfluentAdminClientInterface = &kafka.AdminClient{}

// Create A New Kafka (Confluent, etc...) AdminClient Based On The Kafka Secret In The Specified K8S Namespace
func NewKafkaAdminClient(logger *zap.Logger, k8sNamespace string) (AdminClientInterface, error) {

	// Get A List Of The Kafka Secrets From The K8S Namespace
	k8sClient := GetKubernetesClientWrapper(logger)
	kafkaSecrets, err := util2.GetKafkaSecrets(k8sClient, k8sNamespace)
	if err != nil {
		logger.Error("Failed To Get Kafka Secrets", zap.String("Namespace", k8sNamespace), zap.Error(err))
		return nil, err
	}

	// Should Only Be One Kafka Secret In The Namespace
	if len(kafkaSecrets.Items) != 1 {
		logger.Error(fmt.Sprintf("Expected 1 Kafka Secret But Found %d", len(kafkaSecrets.Items)))
		err = errors.New(fmt.Sprintf("expected 1 Kafka Secret but found %d", len(kafkaSecrets.Items)))
		return nil, err
	}

	// Extract The Relevant Data From The Kafka Secret
	kafkaSecret := kafkaSecrets.Items[0]
	brokers := string(kafkaSecret.Data[constants.KafkaSecretKeyBrokers])
	username := string(kafkaSecret.Data[constants.KafkaSecretKeyUsername])
	password := string(kafkaSecret.Data[constants.KafkaSecretKeyPassword])

	// Validate Secret Data
	if !validateKafkaSecret(logger, &kafkaSecret) {
		err = errors.New("invalid Kafka Secret found")
		return nil, err
	}

	// Create The Kafka Consumer Configuration
	configMap := getBaseAdminConfigMap(brokers)

	// Append SASL Authentication If Specified
	if len(username) > 0 && len(password) > 0 {
		util.AddSaslAuthentication(configMap, constants.ConfigPropertySaslMechanismsPlain, username, password)
	}

	// Create A New Kafka AdminClient From ConfigMap & Return Results
	adminClient, err := kafka.NewAdminClient(configMap)
	if err != nil {
		logger.Error("Failed To Create Kafka AdminClient", zap.Any("ConfigMap", configMap), zap.Error(err))
		return nil, err
	}

	// Create And Return A New Kafka AdminClient With Admin Client
	return &KafkaAdminClient{
		logger:      logger,
		namespace:   k8sNamespace,
		kafkaSecret: kafkaSecret.Name,
		adminClient: adminClient,
	}, nil
}

// CreateTopics - Confluent Pass-Through Function
func (c KafkaAdminClient) CreateTopics(ctx context.Context, topicSpecifications []kafka.TopicSpecification, options ...kafka.CreateTopicsAdminOption) ([]kafka.TopicResult, error) {
	return c.adminClient.CreateTopics(ctx, topicSpecifications, options...)
}

// DeleteTopics - Confluent Pass-Through Function
func (c KafkaAdminClient) DeleteTopics(ctx context.Context, topics []string, options ...kafka.DeleteTopicsAdminOption) ([]kafka.TopicResult, error) {
	return c.adminClient.DeleteTopics(ctx, topics, options...)
}

// Close - Confluent Pass-Through Function
func (c KafkaAdminClient) Close() {
	c.adminClient.Close()
}

// Get The K8S Secret With Kafka Credentials For The Specified Topic
func (c KafkaAdminClient) GetKafkaSecretName(topicName string) string {
	return c.kafkaSecret
}

// Utility Function For Returning The Base/Common Kafka ConfigMap (Values Shared By All Connections)
func getBaseAdminConfigMap(brokers string) *kafka.ConfigMap {
	return &kafka.ConfigMap{
		constants.ConfigPropertyBootstrapServers: brokers,
		constants.ConfigPropertyRequestTimeoutMs: constants.ConfigPropertyRequestTimeoutMsValue,
	}
}

// Utility Function For Validating Kafka Secret
func validateKafkaSecret(logger *zap.Logger, secret *corev1.Secret) bool {

	// Assume Invalid Until Proven Otherwise
	valid := false

	// Validate The Kafka Secret
	if secret != nil {

		// Extract The Relevant Data From The Kafka Secret
		brokers := string(secret.Data[constants.KafkaSecretKeyBrokers])
		username := string(secret.Data[constants.KafkaSecretKeyUsername])
		password := string(secret.Data[constants.KafkaSecretKeyPassword])

		// Validate Kafka Secret Data
		if len(brokers) > 0 && len(username) > 0 && len(password) > 0 {

			// Mark Kafka Secret As Valid
			valid = true

		} else {

			// Invalid Kafka Secret - Log State
			pwdString := ""
			if len(password) > 0 {
				pwdString = "********"
			}
			logger.Error("Kafka Secret Contains Invalid Data",
				zap.String("Name", secret.Name),
				zap.String("Brokers", brokers),
				zap.String("Username", username),
				zap.String("Password", pwdString))
		}
	}

	// Return Kafka Secret Validity
	return valid
}
