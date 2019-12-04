package kafkasubscription

import (
	"context"
	"fmt"
	eventingduck "github.com/knative/eventing/pkg/apis/duck/v1alpha1"
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"github.com/kyma-incubator/knative-kafka/components/controller/constants"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/env"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/event"
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/util"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// Define A Kafka Subscription Reconciler
type Reconciler struct {
	client      client.Client
	recorder    record.EventRecorder
	logger      *zap.Logger
	adminClient kafkaadmin.AdminClientInterface
	environment *env.Environment
}

// Create A New "Kafka" (Knative) Subscription Reconciler
func NewReconciler(recorder record.EventRecorder, logger *zap.Logger, adminClient kafkaadmin.AdminClientInterface, environment *env.Environment) *Reconciler {
	return &Reconciler{
		recorder:    recorder,
		logger:      logger,
		adminClient: adminClient,
		environment: environment,
	}
}

// Verify The Reconciler Implements The Client & Reconciler Interface
var _ inject.Client = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// Implementation The Client Interface
func (r *Reconciler) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

//
// Implement The Reconciler Interface
//
// Reconcile the current Subscription state with the new/desired state
//
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Append The Request To The Logger
	logger := r.logger.With(zap.Any("Request", request))

	// Get The Subscription From The Request
	ctx := context.TODO()
	subscription := &eventingv1alpha1.Subscription{}
	err := r.client.Get(ctx, request.NamespacedName, subscription)

	// If Subscription Was Deleted Prior To This There Is Nothing To Do
	//   (This is NOT the normal Subscription deletion workflow and shouldn't happen due to finalizers.)
	if errors.IsNotFound(err) {
		logger.Info("Subscription Not Found - Skipping")
		return reconcile.Result{}, nil
	}

	// Other Errors Should Be Retried
	if err != nil {
		logger.Error("Failed To Get Subscription - Retrying", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Verify The Subscription Is For A KafkaChannel
	isKafkaChannel, err := r.isKafkaChannel(ctx, subscription)
	if err != nil {
		// If The Indirect Knative Messaging Channel Was Not Found
		if errors.IsNotFound(err) {
			if subscription.DeletionTimestamp == nil {
				// Subscription Is NOT Being Deleted - Requeue
				logger.Warn("Unable To Determine Whether Subscription Is For Kafka Channel - Requeuing", zap.Any("Channel", subscription.Spec.Channel))
				return reconcile.Result{Requeue: true}, nil
			} else {
				// Subscription IS Being Deleted - Continue
				logger.Info("Unable To Determine Whether Deleted Subscription Is For Kafka Channel - Continuing", zap.Any("Channel", subscription.Spec.Channel))
			}
		} else {
			logger.Error("Failed To Determine Whether Subscription Is For Kafka Channel - Requeuing", zap.Any("Channel", subscription.Spec.Channel), zap.Error(err))
			return reconcile.Result{Requeue: true}, nil
		}
	} else if !isKafkaChannel {
		logger.Debug("Subscription Is Not For Kafka Channel - Skipping", zap.Any("Channel", subscription.Spec.Channel))
		return reconcile.Result{}, nil
	}

	// Deep Copy The Subscription (Immutability ;)
	subscription = subscription.DeepCopy()

	// Get The KafkaChannel Associated With This Subscription (Likely Already Torn Down In Delete DataFlow Use Case)
	objectKey := client.ObjectKey{Namespace: subscription.Namespace, Name: subscription.Spec.Channel.Name} // TODO - This relies on the knative messaging channel and the kafkachannel having the same name!
	channel := &kafkav1alpha1.KafkaChannel{}
	err = r.client.Get(ctx, objectKey, channel)
	if err != nil {
		if errors.IsNotFound(err) {
			// If The Channel IS Non-Existent And Subscription Is NOT Being Deleted - Log Warning & Requeue
			// (This shouldn't happen - not sure what a stranded subscription should do anyway?)
			if subscription.DeletionTimestamp == nil {
				logger.Warn("Encountered Non-Deletion Subscription With Missing Channel - Retrying")
				return reconcile.Result{Requeue: true}, nil
			} else {
				// Subscription IS Being Deleted And Channel Was ALREADY Deleted - Reset Channel To Nil
				channel = nil
			}
		} else {
			logger.Error("Failed To Get Channel For Subscription - Retrying", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	// If The Channel Was Found Then Deep Copy It
	if channel != nil {
		channel = channel.DeepCopy()
	}

	// If Finalizer Is Missing & Subscription Is Not Being Deleted - Add Finalizer To The Subscription & Requeue
	if subscription.DeletionTimestamp == nil {
		addFinalizerResult := util.AddFinalizerToSubscription(subscription, constants.KafkaSubscriptionControllerAgentName)
		if addFinalizerResult == util.FinalizerAdded {
			logger.Info("Added Finalizer To Subscription", zap.String("Subscription.Namespace", subscription.Namespace), zap.String("Subscription.Name", subscription.Name))
			err = r.updateSubscriptionFinalizers(ctx, subscription)
			return reconcile.Result{Requeue: true}, err
		}
	}

	// Perform The Subscription Reconciliation
	logger.Info("Subscription Owned By Controller - Reconciling", zap.Any("Subscription.Spec", subscription.Spec))
	err = r.reconcile(ctx, subscription, channel)
	if err != nil {
		logger.Error("Failed To Reconcile Subscription", zap.Error(err))
		// Note: Do NOT Return Error Here In Order To Ensure Status Update
	} else {
		logger.Info("Successfully Reconciled Subscription", zap.Any("Subscription", subscription))
	}

	// If The Channel Exists Then Update The Channel SubscribableStatus & Return Results
	if channel != nil {
		updateStatusErr := r.updateChannelStatus(ctx, channel)
		if updateStatusErr != nil {
			r.recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelUpdateFailed.String(), "Failed To Update Channel Status: %v", err)
			logger.Error("Failed To Update Channel Status", zap.Error(updateStatusErr))
			return reconcile.Result{}, updateStatusErr
		}
	}

	// Return Reconcile Results
	return reconcile.Result{}, err
}

// Perform The Actual Subscription Reconciliation
func (r *Reconciler) reconcile(ctx context.Context, subscription *eventingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel) error {

	// Create The Reconciliation Completion Channels
	reconcileDispatcherCompleteChannel := make(chan error)

	// Reconcile The Various Components In Parallel
	go r.reconcileDispatcher(ctx, subscription, channel, reconcileDispatcherCompleteChannel)

	// Block On All Sub-Reconciliation Go Routine Completion
	reconcileDispatcherErr := <-reconcileDispatcherCompleteChannel

	// Handle All Reconciliation Errors - Log & Return
	if reconcileDispatcherErr != nil {
		r.updateSubscribableStatus(ctx, subscription, channel, corev1.ConditionFalse)
		return fmt.Errorf("reconciliation failed")
	}

	// If The Subscription Is Being Deleted Then Remove The Finalizer So That The Subscription CR Can Be Deleted
	// (Relying On K8S Garbage Collection To Delete Subscription Deployment, Service Based On Subscription OwnerReference)
	if subscription.DeletionTimestamp != nil {
		r.logger.Info("Subscription Deleted - Removing Finalizer", zap.Time("DeletionTimestamp", subscription.DeletionTimestamp.Time))
		removeFinalizerResult := util.RemoveFinalizerFromSubscription(subscription, constants.KafkaSubscriptionControllerAgentName)
		if removeFinalizerResult == util.FinalizerRemoved {
			err := r.updateSubscriptionFinalizers(ctx, subscription)
			return err
		}
	}

	// Update The Channel's SubscribableStatus
	r.updateSubscribableStatus(ctx, subscription, channel, corev1.ConditionTrue)

	// Return Success
	return nil
}

// Determine Whether The Channel Associated With The Specified Subscription Is A KafkaChannel
func (r *Reconciler) isKafkaChannel(ctx context.Context, subscription *eventingv1alpha1.Subscription) (bool, error) {

	// If The Subscription Is Directly For A KafkaChannel
	if subscription.Spec.Channel.APIVersion == kafkav1alpha1.SchemeGroupVersion.String() && subscription.Spec.Channel.Kind == constants.KafkaChannelKind {

		// Then Nothing More To Do - Return True
		return true, nil

	} else {

		// Otherwise If The Subscription Is For A Knative Messaging Channel
		if subscription.Spec.Channel.APIVersion == messagingv1alpha1.SchemeGroupVersion.String() && subscription.Spec.Channel.Kind == constants.KnativeChannelKind {

			// Then Attempt To Load The Knative Messaging Channel
			objectKey := client.ObjectKey{Namespace: subscription.Namespace, Name: subscription.Spec.Channel.Name}
			messagingChannel := &messagingv1alpha1.Channel{}
			err := r.client.Get(ctx, objectKey, messagingChannel)
			if err != nil {

				// Return Any Errors (Including NotFound)
				return false, err

			} else {

				// Return Results Based On Whether The Indirect Channel Is A KafkaChannel
				if messagingChannel.Spec.ChannelTemplate.APIVersion == kafkav1alpha1.SchemeGroupVersion.String() && messagingChannel.Spec.ChannelTemplate.Kind == constants.KafkaChannelKind {
					return true, nil
				} else {
					return false, nil
				}
			}

		} else {

			// Subscription Is Not For KafkaChannel Either Direct/Indirect
			return false, nil
		}
	}
}

// Update The Subscription Finalizers / Status & Push To K8S (Lifted From Knative Channel Util)
func (r *Reconciler) updateSubscriptionFinalizers(ctx context.Context, subscription *eventingv1alpha1.Subscription) error {

	// Get Subscription Specific Logger
	logger := util.SubscriptionLogger(r.logger, subscription)

	// Get The Current Subscription From K8S
	objectKey := client.ObjectKey{Namespace: subscription.Namespace, Name: subscription.Name}
	k8sSubscription := &eventingv1alpha1.Subscription{}
	err := r.client.Get(ctx, objectKey, k8sSubscription)
	if err != nil {
		logger.Error("Failed To Get Subscription To Update Finalizers", zap.Error(err))
		return err
	}

	// Compare & Update Finalizers If Different While Tracking Changes
	if !equality.Semantic.DeepEqual(k8sSubscription.Finalizers, subscription.Finalizers) {
		k8sSubscription = k8sSubscription.DeepCopy()
		k8sSubscription.SetFinalizers(subscription.ObjectMeta.Finalizers)
		err := r.client.Update(ctx, k8sSubscription)
		if err != nil {
			logger.Error("Failed To Update Subscription Finalizers", zap.Error(err))
			return err
		}
	}

	// Return Success
	return nil
}

// Update The Channel's Subscribable Status & Push To K8S
func (r *Reconciler) updateSubscribableStatus(ctx context.Context, subscription *eventingv1alpha1.Subscription, channel *kafkav1alpha1.KafkaChannel, status corev1.ConditionStatus) {

	// Get Subscription Specific Logger
	logger := util.SubscriptionLogger(r.logger, subscription)

	// Validate The Channel And It's Subscribables
	if channel == nil {
		logger.Info("Unable To Update SubscribableStatus - Channel Is Nil (likely already deleted)")
		return
	} else if channel.Spec.Subscribable == nil || len(channel.Spec.Subscribable.Subscribers) <= 0 {
		logger.Warn("Unable To Update SubscribableStatus - Channel Has No Subscribers")
		return
	}

	// Get The Channels' SubscribableStatus (An Array Of SubscriberStatus Entries)
	subscribableStatus := channel.Status.SubscribableTypeStatus.SubscribableStatus

	// Initialize The Channels' SubscribableStatus If Necessary
	if subscribableStatus == nil || subscribableStatus.Subscribers == nil {
		subscribableStatus = &eventingduck.SubscribableStatus{
			Subscribers: make([]eventingduck.SubscriberStatus, 0),
		}
		channel.Status.SubscribableTypeStatus.SubscribableStatus = subscribableStatus
	}

	// Loop Over All The Channel's SubscriberStatus Entries - Update Existing Entry If Found
	found := false
	for _, subscriberStatus := range subscribableStatus.Subscribers {
		if subscriberStatus.UID == subscription.UID {
			subscriberStatus.Ready = status
			found = true
			break
		}
	}

	// If SubscriberStatus Entry Not Found - Then Create New One & Update Channel's SubscribableStatus
	if !found {
		subscriberStatus := eventingduck.SubscriberStatus{
			UID:                subscription.UID,
			ObservedGeneration: subscription.Generation,
			Ready:              status,
		}
		subscribableStatus.Subscribers = append(subscribableStatus.Subscribers, subscriberStatus)
	}
}

// Update K8S With The Specified Channel's Status
func (r *Reconciler) updateChannelStatus(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// Get Channel Specific Logger
	logger := util.ChannelLogger(r.logger, channel)

	// Get The Current Channel From K8S
	objectKey := client.ObjectKey{Namespace: channel.Namespace, Name: channel.Name}
	k8sChannel := &kafkav1alpha1.KafkaChannel{}
	if err := r.client.Get(ctx, objectKey, k8sChannel); err != nil {
		logger.Error("Failed To Get Channel From K8S", zap.Error(err))
		return err
	}

	// Compare The Channels Status & Update If Necessary
	if !equality.Semantic.DeepEqual(k8sChannel.Status, channel.Status) {

		// DeepCopy & Update The K8S Channel's Status
		k8sChannel = k8sChannel.DeepCopy()
		k8sChannel.Status = channel.Status

		// Update The Channel's Status SubResource
		if err := r.client.Status().Update(ctx, k8sChannel); err != nil {
			logger.Error("Failed To Update Channel Status", zap.Error(err))
			return err
		}
	}

	// Return Success
	return nil
}
