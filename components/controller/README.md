# Kafka-Channel-Controller

The kafka-channel-controller application is the Kubernetes Controller for the Knative Channel CustomResource.  
The Channel Controller is responsible for creating the Kafka Topic with specified configuration, as well as for 
creating the singular (per-topic) Channel Deployment and Services, as well as the per-Subscription Dispatcher Deployments.  


## Makefile Targets
The Makefile should be used to build and deploy the **kafka-channel-controller** application as follows...

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
  # Build Native Binary
  make build-native
  
  # Build Linux Binary
  make build-linux
  ``` 

- **Test**
  ```
  # Run All Unit Tests
  make test
  ```
    
- **Docker Build & Push**
  ```
  # Build Docker Container (Without Unit Tests)
  make docker-build

  # Build Docker Container (With Unit Tests)
  make BUILD_TESTS=true docker-build
  
  # Push Docker Continer
  make docker-push
  ```

