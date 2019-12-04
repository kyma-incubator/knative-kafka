# Knative-Kafka-Dispatcher

The Knative-Kafka architecture includes a distinct instance of a Kafka Consumer Group for every Subscription
to a Knative Channel (Kafka Topic) in order to achieve it's scalability goals.  The knative-kafka-dispatcher
application is that Consumer Group.  From the Kafka perspective this code is a "consumer" of Kafka messages.
The implementation forwards the messages (in order) to configured the Subscriber URI.

Instances of this application are deployed by the knative-kafka-controller upon detection of changes to a
Channel's spec.Subscribable list of Subscriptions.  The implementation uses the [Confluent Go Client](https://github.com/confluentinc/confluent-kafka-go) 
, and the underlying C/C++ [librdkafka](https://github.com/edenhill/librdkafka) library, to implement the Kafka 
consumer and manage re-balancing of, and connections to, a Kafka cluster.


## Makefile Targets
The Makefile should be used to build and deploy the **knative-kafka-channel** application as follows...

- **Setup**
  ```
  source ./local-env.sh
  ```
  
- **Cleaning**
  ```
  # Clean Application Build
  make clean
  ```

- **Dependencies**
  ```
  # Verify / Dowload Dependencies
  make dep
  ```

- **Building**
  ```
  # Build Individual Native Binaries Into The build/image Directory
  make build-native
  
  # Build Individual Linux Binaries Into The build/image Directory
  make build-linux
  ``` 

- **Test**
  ```
  # Run All Unit Tests
  make test
  ```
    
- **Docker Build & Push**
  ```
  # Build Docker Container
  make docker-build
  
  # Push Docker Continer
  make docker-push
  ```
