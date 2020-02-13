# Knative Kafka

This project is a Knative Eventing implementation of a Kafka backed channel which provides advanced functionality and
production grade qualities.

The following components comprise the total **Knative-Kafka** solution...

- [knative-kafka-common](./components/common/README.md) - Common code used by the other Knative-Kafka components.

- [knative-kafka-channel](./components/channel/README.md) - The event receiver of the Channel to which inbound
messages are sent.  An http server which accepts messages that conform to the CloudEvent specification and then writes
those messages to the corresponding Kafka Topic. This is the "Producer" from the Kafka perspective.
    
- [knative-kafka-controller](./components/controller/README.md) - This component implements the channel controller.  It
was generated with Kubebuilder and includes the `KafkaChannel` CRD. This should not be confused with the 
Knative Eventing Contrib Kafka Channel of which you should only install one.

- [knative-kafka-dispatcher](./components/dispatcher/README.md) - This component runs the Kafka ConsumerGroups
responsible for processing messages from the corresponding Kafka Topic.  This is the "Consumer" from the Kafka
perspective.

## Installation

For installation and configuration instructions please see the helm chart [README](./resources/README.md)

## Support Tooling

The following are one time install of required tools
```
# Install Go (If not already installed)
brew install go

# Install Dep (If not already installed)
brew install dep
brew upgrade dep

# Install Confluent Go Client Dependencies
brew install pkg-config
brew install librdkafka

# If the current version of librdkafka from brew doesn't work with the knative-kafka components then you can roll back to an older version via the following.
curl -LO https://github.com/edenhill/librdkafka/archive/v1.0.1.tar.gz \
  && tar -xvf v1.0.1.tar.gz \
  && librdkafka-1.0.1 \
  && ./configure \
  && make \
  && sudo make install
```
## Control Plane

The control plane for the Kafka Channels is managed by the [knative-kafka-controller](./components/controller/README.md)
which is installed by default into the knative-eventing namespace. `KafkaChannel` custom resource instances can be created in any user
namespace. The knative-kafka-controller will guarantee that the Data Plane is configured to support the flow of events as
defined by [Subscriptions](https://knative.dev/docs/reference/eventing/#messaging.knative.dev/v1alpha1.Subscription) to
a KafkaChannel.  The underlying Kafka infrastructure to be used is defined in a specially
labeled [secret](./resources/README.md#Credentials) in the knative-eventing namespace.  Knative-Kafka supports several
different Kafka (and Kafka-like) [infrastructures](./resources/README.md#Kafka%20Providers).

## Data Plane

The data plane for all `KafkaChannels` runs in the knative-eventing namespace.  There is a single deployment for the
receiver side of all channels which accepts CloudEvents and sends them to Kafka.  Each `KafkaChannel` uses one Kafka topic.
This deployment supports horizontal scaling with linearly increasing performance characteristics through specifying the number of replicas.

Each `KafkaChannel` has one deployment for the dispatcher side which reads from the Kafka topic and sends to subscribers.
Each subscriber has its own Kafka consumer group. This deployment can be scaled up to a replica count equalling the
number of partitions in the Kafka topic.

## Messaging Guarantees

An event sent to a `KafkaChannel` is guaranteed to be persisted and processed if a 202 response is received by the sender.
Events are partitioned in Kafka by the Cloud Event attribute `subject` when present or randomly when absent.  Events
in each partition are processed in order, with an at least once guarantee. If a full cycle of retries for a given
subscription fails, the event is ignored and processing continues with the next event.