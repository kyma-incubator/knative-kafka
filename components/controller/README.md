# Kafka-Channel Controller

The Controller component implements the KafkaChannel CRD (api, client, reconciler, 
etc.) based on the latest knative-eventing SharedMain reconciler framework and
utilities.  It was previously based on legacy kubebuilder / controller-runtime
logic.

**Note** - The controller separation describe in the next paragraph is in-progress
and should be available soon ; )  Currently there is only one reconciler.

The controller defines the KafkaChannel CRD type and reconciles all such instances
in the K8S Cluster.  It actually consists of two reconcilers, one for watching
"Kafka" Secrets (those in knative-eventing labelled 
`knativekafka.kyma-project.io/kafka-secret: "true"`) which provisions the Kafka
Topic and creates the Channel / Producer Deployment/Services and another which 
is watching KafkaChannel resources and creates the Dispatcher / Consumer 
Deployments.

**Note** - Deleting a KafkaChannel CRD instance is destructive in that it will
Remove the Kafka Topic resulting in the loss of all events therein.  While the
Dispatcher and Producer will perform semi-graceful shutdown there is no attempt
to "drain" the topic or complete incoming CloudEvents.

All (most) runtime components are created in the knative-eventing namespace
which presents some challenges in using default K8S OwnerReferences and
Garbage Collection.  Further work is planned to sort out some of the cross-name
issues by refactoring the controller and by using the informers/listers directly.
