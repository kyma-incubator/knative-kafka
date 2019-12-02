# Knative-Kafka-Channel

The Knative-Kafka architecture contains a distinct service endpoint for every Channel / Kafka Topic in
order to achieve a scalable solution.  The knative-kafka-channel application is that endpoint which
simply forwards request body content from POST requests to a Kafka Broker and Topic.  From the Kafka
perspective this code is a "producer" of Kafka messages.

Instances of this application are deployed by the [knative-kafka-controller](../controller/README.md) upon creation 
of knative-kafka Channel CustomResources. The basic AsyncProducer initialization and usage was lifted from early 
versions of Knative Eventing implementations of the Kafka Bus.  The [Conluent Go Client](https://github.com/confluentinc/confluent-kafka-go)
library, along with the underlying C/C++ [librdkafka](https://github.com/edenhill/librdkafka) is used to connect with
one of Kafka brokers specified in the environment variables. 


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
