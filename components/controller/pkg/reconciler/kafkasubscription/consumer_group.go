package kafkasubscription

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaconsumer "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/consumer"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

//
// Reconcile The Kafka ConsumerGroup Offset For The Specified Subscription / Channel
//
func (r *Reconciler) reconcileConsumerGroupOffsets(ctx context.Context,
	subscription *messagingv1alpha1.Subscription,
	channel *kafkav1alpha1.KafkaChannel) error {

	// Get Subscription Specific Logger
	logger := util.SubscriptionLogger(r.logger, subscription)

	// If The Subscription Is Being Deleted - Nothing To Do Since K8S Garbage Collection Will Cleanup Based On OwnerReference
	if subscription.DeletionTimestamp != nil {
		logger.Info("Successfully Reconciled Dispatcher Deletion")
		return nil
	}

	// Get The EventStartTime From Subscription
	subscriptionEventStartTime := util.EventStartTime(subscription, logger)

	// If There Is An Existing Deployment Handle Case Where EventStartTime Changes
	existingDeployment, deployErr := r.getK8sDispatcherDeployment(ctx, subscription)
	if deployErr == nil {

		dispatcherEventStartTime := getEventStartTimeForDispatcher(existingDeployment)

		if dispatcherEventStartTime != subscriptionEventStartTime {
			logger.Info("Resetting Consumer Group Offsets")

			// Delete The Existing Dispatcher Deployment (Shutting Down Dispatchers)
			deleteErr := r.client.Delete(ctx, existingDeployment)
			if deleteErr != nil {
				r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", deleteErr)
			} else {
				logger.Info("Deleted Dispatcher Deployment, To Reset Consumer Group Offsets")
			}

			dateTime, parseErr := time.Parse(time.RFC3339, subscriptionEventStartTime)
			if parseErr != nil {
				logger.Error("Failed To Parse EventStartTime", zap.Error(parseErr))
				r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", parseErr)
			}

			// Reset The Consumer Group Offsets To The Given EventStartTime
			resetErr := r.setConsumerGroupOffsets(ctx, subscription, channel, dateTime)
			if resetErr != nil {
				r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", resetErr)
			} else {
				logger.Info("Successfully Reset Consumer Group Offsets For Dispatcher")
			}

			return resetErr
		}
		// If There Is No Dispatcher, Set The Offsets
	} else if errors.IsNotFound(deployErr) {
		dateTime, parseErr := time.Parse(time.RFC3339, subscriptionEventStartTime)

		if parseErr != nil {
			logger.Error("Failed To Parse EventStartTime", zap.Error(parseErr))
			r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", parseErr)
		}

		// Reset The Consumer Group Offsets To The Given EventStartTime
		resetErr := r.setConsumerGroupOffsets(ctx, subscription, channel, dateTime)
		if resetErr != nil {
			r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Reconcile K8S Deployment For Dispatcher: %v", resetErr)
		} else {
			logger.Info("Successfully Reset Consumer Group Offsets For Dispatcher")
		}

		return resetErr
	}

	logger.Info("Not Necessary To Reset Consumer Group Offsets")
	return nil
}

// Reset The Consumer Group Offsets For a Given Consumer Group To The Offset Closest To The Specified Time
func (r *Reconciler) setConsumerGroupOffsets(ctx context.Context, subscription *messagingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel, dateTime time.Time) error {

	// Get The GroupId
	groupId := subscription.Name

	// Get The Channels Topic Name
	topicName := util.TopicName(channel, r.environment)

	// Setup The Logger
	logger := r.logger.With(zap.String("Topic", topicName), zap.String("Group", groupId))

	// Get The Kafka Secret From K8S
	kafkaSecret, kafkaSecretErr := r.getKafkaSecret(ctx, channel.Namespace, topicName)
	if kafkaSecretErr != nil {
		logger.Error("Failed To Get Kafka Secret", zap.Error(kafkaSecretErr))
		r.recorder.Eventf(subscription, corev1.EventTypeWarning, event.DispatcherDeploymentReconciliationFailed.String(), "Failed To Get Kafka Secret For Channel: %v", kafkaSecretErr)
	}

	// Extract The Kafka Secret Data
	kafkaBrokers := string(kafkaSecret.Data[constants.KafkaSecretDataKeyBrokers])
	kafkaUsername := string(kafkaSecret.Data[constants.KafkaSecretDataKeyUsername])
	kafkaPassword := string(kafkaSecret.Data[constants.KafkaSecretDataKeyPassword])

	// Consumer Group Manipulation Is Part Of The Kafka Consumer API
	consumer, err := kafkaconsumer.CreateConsumer(kafkaBrokers, groupId, "latest", kafkaUsername, kafkaPassword)
	if err != nil {
		logger.Error("Unable to connect to kafka", zap.Error(err))
		return err
	}

	// Close The Consumer When Finished
	defer safeConsumerClose(logger, consumer)

	// Retrieve The Partition Metadata For The Topic
	metadata, err := consumer.GetMetadata(&topicName, false, KafkaApiTimeout)
	if err != nil {
		logger.Error("Unable to retrieve topic metadata", zap.Error(err))
		return err
	}

	// Build Slice Of TopicPartition Structs Based On Topic Metadata
	topicPartitions := make([]kafka.TopicPartition, 0)
	for _, partitionMetadata := range metadata.Topics[topicName].Partitions {
		offset := kafka.Offset(dateTime.UnixNano() / 1000000)
		topicPartitions = append(topicPartitions, kafka.TopicPartition{Partition: partitionMetadata.ID, Topic: &topicName, Offset: offset})
	}

	// Get The Offsets For All Partitions For The Given Date/Time
	topicPartitions, err = consumer.OffsetsForTimes(topicPartitions, KafkaApiTimeout)
	if err != nil {
		logger.Error("Can't get offsets for times", zap.Error(err))
		return err
	}

	// Convert Relative Offsets (End, Beginning etc.) To Actual Numeric Offsets
	convertRelativeOffsets(consumer, topicPartitions)

	// Commit The Consumer Group Offsets To The Retrieved Offsets
	topicPartitions, err = consumer.CommitOffsets(topicPartitions)
	if err != nil && err.(kafka.Error).Code() != kafka.ErrNoOffset {
		logger.Warn("Can't commit offsets, this is likely due to the deployment still shutting down", zap.Error(err))
		return err
	}

	return nil
}

// Get The Kafka Secret For The Specified Channel
func (r *Reconciler) getKafkaSecret(ctx context.Context, namespace string, topicName string) (*corev1.Secret, error) {

	// Get The Kafka Secret Associated With The Channel
	kafkaSecretName := r.adminClient.GetKafkaSecretName(topicName)

	// Get The Kafka Secret
	objectKey := client.ObjectKey{Namespace: namespace, Name: kafkaSecretName}
	kafkaSecret := &corev1.Secret{}
	err := r.client.Get(ctx, objectKey, kafkaSecret)
	if err != nil {
		r.logger.Error("Failed To Get Kafka Secret", zap.Any("ObjectKey", objectKey), zap.Error(err))
		return nil, err
	}

	// Return Success
	return kafkaSecret, nil
}

//
// Utilities
//

// Get The Event Start Time From the Dispatcher Deployment
func getEventStartTimeForDispatcher(deployment *appsv1.Deployment) string {
	return deployment.ObjectMeta.Annotations[EventStartTimeAnnotation]
}

// Convert Relative Offsets To Actual Numeric Values
func convertRelativeOffsets(consumer kafkaconsumer.ConsumerInterface, topicPartitions []kafka.TopicPartition) {
	for i := 0; i < len(topicPartitions); i++ {
		if topicPartitions[i].Offset == kafka.OffsetEnd {
			_, high, _ := consumer.QueryWatermarkOffsets(*topicPartitions[i].Topic, topicPartitions[i].Partition, KafkaApiTimeout)
			topicPartitions[i].Offset = kafka.Offset(high)
		} else if (topicPartitions)[i].Offset == kafka.OffsetBeginning {
			topicPartitions[i].Offset = 0
		}
	}
}

// "Safe" Consumer Close For Use With defer
func safeConsumerClose(logger *zap.Logger, consumer kafkaconsumer.ConsumerInterface) {
	err := consumer.Close()
	if err != nil {
		logger.Error("Failed To Close Kafka Consumer", zap.Error(err))
	}
}
