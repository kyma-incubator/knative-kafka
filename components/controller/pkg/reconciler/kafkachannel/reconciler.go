package kafkachannel

import (
	"context"
	"fmt"
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
	"time"
)

// Define A KafkaChannel Reconciler
type Reconciler struct {
	client      client.Client
	recorder    record.EventRecorder
	logger      *zap.Logger
	adminClient kafkaadmin.AdminClientInterface
	environment *env.Environment
}

// Create A New KafkaChannel Reconciler
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
// Reconcile the current Channel state with the new/desired state & update the Channel's Status block
//
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Append The Request To The Logger
	logger := r.logger.With(zap.Any("Request", request))
	logger.Debug("<==========  START CHANNEL RECONCILIATION  ==========>")

	// Get The Channel From The Request
	ctx := context.TODO()
	channel := &kafkav1alpha1.KafkaChannel{}
	err := r.client.Get(ctx, request.NamespacedName, channel)

	// If Channel Was Deleted Prior To This There Is Nothing To Do
	//   (This is NOT the normal Channel deletion workflow and shouldn't happen due to finalizers.)
	if errors.IsNotFound(err) {
		logger.Info("Channel Not Found - Skipping")
		return reconcile.Result{}, nil
	}

	// Other Errors Should Be Retried
	if err != nil {
		logger.Error("Failed To Get Channel - Retrying", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Deep Copy The Channel (Immutability ;)
	channel = channel.DeepCopy()

	// Add Finalizer To The Channel & Requeue If First Time
	addFinalizerResult := util.AddFinalizerToChannel(channel, constants.KafkaChannelControllerAgentName)
	if addFinalizerResult == util.FinalizerAdded {
		updateFinalizersErr := r.updateChannelFinalizers(ctx, channel)
		if updateFinalizersErr != nil {
			r.recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelUpdateFailed.String(), "Failed To Update Channel Finalizer: %v", updateFinalizersErr)
			logger.Error("Failed To Update Finalizers On Channel", zap.Error(updateFinalizersErr))
			return reconcile.Result{}, updateFinalizersErr
		} else {
			logger.Info("Successfully Added Finalizer To Channel", zap.String("Channel.Namespace", channel.Namespace), zap.String("Channel.Name", channel.Name))
			return reconcile.Result{Requeue: true}, nil
		}
	}

	// Reset The Channel's Status Conditions To Unknown (Addressable, Topic, Service, Deployment, etc...)
	channel.Status.InitializeConditions()

	// Perform The Channel Reconciliation
	logger.Info("Channel Owned By Controller - Reconciling", zap.Any("Channel.Spec", channel.Spec))
	err = r.reconcile(ctx, channel)
	if err != nil {
		logger.Error("Failed To Reconcile Channel", zap.Error(err))
		// Note: Do NOT Return Error Here In Order To Ensure Status Update
	} else {
		logger.Info("Successfully Reconciled Channel", zap.Any("Channel", channel))
	}

	// If NOT Being Deleted - Update The Channel Reconciliation Status & Return Results
	if channel.DeletionTimestamp == nil {
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

// Perform The Actual Channel Reconciliation (In Parallel)
func (r *Reconciler) reconcile(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// NOTE - The sequential order of reconciliation must be "Topic" then "Channel / Dispatcher" in order for the
	//        EventHub Cache to know the dynamically determined EventHub Namespace / Kafka Secret selected for the topic.

	// Reconcile The KafkaChannel's Kafka Topic
	err := r.reconcileTopic(ctx, channel)
	if err != nil {
		return fmt.Errorf("reconciliation failed")
	}

	// Reconcile The KafkaChannel's Channel & Dispatcher Deployment/Service
	channelError := r.reconcileChannel(ctx, channel)
	dispatcherError := r.reconcileDispatcher(ctx, channel)
	if channelError != nil || dispatcherError != nil {
		return fmt.Errorf("reconciliation failed")
	}

	// If The Channel Is Being Deleted Then Remove The Finalizer So That The Channel CR Can Be Deleted
	// (Relying On K8S Garbage Collection To Delete Channel Deployment, Service Based On Channel OwnerReference)
	if channel.DeletionTimestamp != nil {
		r.logger.Info("Channel Deleted - Removing Finalizer", zap.Time("DeletionTimestamp", channel.DeletionTimestamp.Time))
		util.RemoveFinalizerFromChannel(channel, constants.KafkaChannelControllerAgentName)
		updateFinalizersErr := r.updateChannelFinalizers(ctx, channel)
		if updateFinalizersErr != nil {
			r.recorder.Eventf(channel, corev1.EventTypeWarning, event.ChannelUpdateFailed.String(), "Failed To Update Channel Finalizer: %v", updateFinalizersErr)
			r.logger.Error("Failed To Remove Finalizers On Channel", zap.Error(updateFinalizersErr))
			return updateFinalizersErr
		} else {
			r.logger.Info("Successfully Removed Finalizer From Channel", zap.String("Channel.Namespace", channel.Namespace), zap.String("Channel.Name", channel.Name))
			return nil
		}
	}

	// Return Success
	return nil
}

// Update K8S With The Specified Channel's Finalizers
func (r *Reconciler) updateChannelFinalizers(ctx context.Context, channel *kafkav1alpha1.KafkaChannel) error {

	// Get Channel Specific Logger
	logger := util.ChannelLogger(r.logger, channel)

	// Get The Current Channel From K8S
	objectKey := client.ObjectKey{Namespace: channel.Namespace, Name: channel.Name}
	k8sChannel := &kafkav1alpha1.KafkaChannel{}
	if err := r.client.Get(ctx, objectKey, k8sChannel); err != nil {
		logger.Error("Failed To Get Channel From K8S", zap.Error(err))
		return err
	}

	// Compare The Channels Finalizers & Update If Necessary
	if !equality.Semantic.DeepEqual(k8sChannel.Finalizers, channel.Finalizers) {
		k8sChannel = k8sChannel.DeepCopy()
		k8sChannel.SetFinalizers(channel.ObjectMeta.Finalizers)
		if err := r.client.Update(ctx, k8sChannel); err != nil {
			logger.Error("Failed To Update Channel Finalizers", zap.Error(err))
			return err
		}
	}

	// Return Success
	return nil
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

	// Determine Whether The Channel Is Becoming "Ready"
	becomingReady := channel.Status.IsReady() && !k8sChannel.Status.IsReady()

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

	// Log The Channel Becoming Ready
	if becomingReady {
		duration := time.Since(k8sChannel.ObjectMeta.CreationTimestamp.Time)
		logger.Info("Channel Is Ready!", zap.Any("Duration", duration))
	}

	// Return Success
	return nil
}
