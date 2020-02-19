package kafkachannel

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

// Reconcile The Kafka Topic Associated With The Specified Channel & Return The Kafka Secret
func (r *Reconciler) reconcileTopic(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// Get The TopicName For Specified Channel
	topicName := util.TopicName(channel)

	// Get Channel Specific Logger & Add Topic Name
	logger := util.ChannelLogger(r.Logger.Desugar(), channel).With(zap.String("TopicName", topicName))

	// Get The Topic Configuration (First From Channel With Failover To Environment)
	numPartitions := util.NumPartitions(channel, r.environment, r.Logger.Desugar())
	replicationFactor := util.ReplicationFactor(channel, r.environment, r.Logger.Desugar())
	retentionMillis := util.RetentionMillis(channel, r.environment, r.Logger.Desugar())

	// Error Reference
	var err error

	{
		// TODO - Similar to the above concerns we'll need requirements around handling changes to a Channel.Spec.Arguments where the partitions or replication is changed?
		// TODO - Will we support that at all?  How? etc...
		// TODO - The kafka command line tool for topics exposes the "alter" option which can be used to change partitions and such - need to check Confluent client.

		// Otherwise Create The Topic (Handles Case Where Already Exists)
		err = r.createTopic(ctx, topicName, numPartitions, replicationFactor, retentionMillis)
	}

	// Log Results & Return Status
	if err != nil {
		r.Recorder.Eventf(channel, corev1.EventTypeWarning, event.KafkaTopicReconciliationFailed.String(), "Failed To Reconcile Kafka Topic For Channel: %v", err)
		logger.Error("Failed To Reconcile Topic", zap.Error(err))
		channel.Status.MarkTopicFailed("TopicFailed", fmt.Sprintf("Channel Kafka Topic Failed: %s", err))
	} else {
		logger.Info("Successfully Reconciled Topic")
		channel.Status.MarkTopicTrue()
	}
	return err
}

// Create The Specified Kafka Topic
func (r *Reconciler) createTopic(ctx context.Context, topicName string, partitions int, replicationFactor int, retentionMillis int64) error {

	// Setup The Logger
	logger := r.Logger.Desugar().With(zap.String("Topic", topicName))

	// Create The TopicSpecification
	topicSpecifications := []kafka.TopicSpecification{
		{
			Topic:             topicName,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
			Config: map[string]string{
				KafkaTopicConfigRetentionMs: strconv.FormatInt(retentionMillis, 10),
			},
		},
	}

	// Attempt To Create The Topic & Process Results
	topicResults, err := r.adminClient.CreateTopics(ctx, topicSpecifications)
	if len(topicResults) > 0 {
		topicResultError := topicResults[0].Error
		topicResultErrorCode := topicResultError.Code()
		if topicResultErrorCode == kafka.ErrTopicAlreadyExists {
			logger.Info("Kafka Topic Already Exists - No Creation Required")
			return nil
		} else if topicResultErrorCode == kafka.ErrNoError {
			logger.Info("Successfully Created New Kafka Topic")
			return nil
		} else {
			logger.Error("Failed To Create Topic (Results)", zap.Error(err), zap.Any("TopicResults", topicResults))
			return topicResults[0].Error
		}
	} else if err != nil {
		logger.Error("Failed To Create Topic (Error)", zap.Error(err))
		return err
	} else {
		logger.Warn("Received Empty TopicResults From CreateTopics Request")
		return nil
	}
}

// Delete The Specified Kafka Topic
func (r *Reconciler) deleteTopic(ctx context.Context, topicName string) error {

	// Setup The Logger
	logger := r.Logger.Desugar().With(zap.String("Topic", topicName))

	// Attempt To Delete The Topic & Process Results
	topicResults, err := r.adminClient.DeleteTopics(ctx, []string{topicName})
	if len(topicResults) > 0 {
		topicResultError := topicResults[0].Error
		topicResultErrorCode := topicResultError.Code()
		logger = logger.With(zap.Any("topicResultErrorCode", topicResultErrorCode))
		if topicResultErrorCode == kafka.ErrUnknownTopic ||
			topicResultErrorCode == kafka.ErrUnknownPartition ||
			topicResultErrorCode == kafka.ErrUnknownTopicOrPart {

			logger.Info("Kafka Topic or Partition Not Found - No Deletion Required")
			return nil
		} else if topicResultErrorCode == kafka.ErrInvalidConfig && r.environment.KafkaProvider == env.KafkaProviderValueAzure {
			// While this could be a valid Kafka error, this most likely is coming from our custom EventHub AdminClient
			// implementation and represents the fact that the EventHub Cache does not contain this topic.  This can
			// happen when an EventHub could not be created due to exceeding the number of allowable EventHubs.  The
			// KafkaChannel is then in an "UNKNOWN" state having never been fully reconciled.  We want to swallow this
			// error here so that the deletion of the Topic / EventHub doesn't block the deletion of the KafkaChannel.
			logger.Warn("Invalid Kafka Topic Configuration (Likely EventHub Namespace Cache) - Unable To Delete", zap.Error(err))
			return nil
		} else if topicResultErrorCode == kafka.ErrNoError {
			logger.Info("Successfully Deleted Existing Kafka Topic")
			return nil
		} else {
			logger.Error("Failed To Delete Topic (Results)", zap.Error(err), zap.Any("TopicResults", topicResults))
			return topicResults[0].Error
		}
	} else if err != nil {
		logger.Error("Failed To Delete Topic (Error)", zap.Error(err))
		return err
	} else {
		logger.Warn("Received Empty TopicResults From DeleteTopics Request")
		return nil
	}
}
