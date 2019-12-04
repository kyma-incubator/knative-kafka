# ###############################################################################
#
#                     Knative-Kafka Channel Makefile
#
# ###############################################################################

# Project Directories
BUILD_ROOT:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR:=$(BUILD_ROOT)/build
BUILD_TESTS:=false

# Application Names (Align With Knative Eventing Version)
APP_NAME=knative-kafka-channel
APP_VERSION=latest

# Build / Packaging / Deployment Configuration
APP_PATH=eu.gcr.io/kyma-project/incubator
APP_TAG=$(APP_NAME):$(APP_VERSION)

#
# Clean Build Image
#

clean:
	@echo 'Cleaning $(APP_NAME) Build Image'
	rm -f $(BUILD_DIR)/$(APP_NAME)

.PHONY: clean

#
# Dep / Fmt / Vet / Lint / Test / Sonar
#

dep:
	@echo 'Ensuring Dependencies'
	cd $(BUILD_ROOT); dep ensure -v

format:
	@echo 'Formatting Go Source'
	go fmt ./...

vet:
	@echo 'Vetting Packages'
	go vet -v ./cmd/...
	go vet -v ./internal/...

lint:
	@echo 'Linting Packages'
	golint ./internal/...

test:
	@echo 'Testing Packages'
	cd $(BUILD_ROOT); mkdir -p $(BUILD_DIR); go test -v ./internal/... -tags static -coverprofile ${BUILD_DIR}/coverage.out
	cd $(BUILD_ROOT); go tool cover -func=${BUILD_DIR}/coverage.out

.PHONY: dep format vet lint test


#
# Build The Binaries
#

build-native: clean
	@echo 'Building $(APP_NAME) (Native)'
	mkdir -p $(BUILD_DIR); cd $(BUILD_ROOT)/cmd; go build -a -tags static -o $(BUILD_DIR)/$(APP_NAME)

build-linux: clean
	@echo 'Building $(APP_NAME) (Linux)'
	mkdir -p $(BUILD_DIR); cd $(BUILD_ROOT)/cmd; CGO_ENABLED=0 GOOS=linux go build -o $(BUILD_DIR)/$(APP_NAME)

.PHONY: build-native build-linux


#
# Package & Push Docker Container
#

docker-build:
	@echo 'Building $(APP_NAME) Docker Container'
	docker build --rm -t $(APP_TAG) --build-arg RUN_TESTS=${BUILD_TESTS} -f ./Dockerfile ../..
	docker tag $(APP_TAG) $(APP_PATH)/$(APP_TAG)

docker-push: docker-build
	@echo 'Pushing $(APP_NAME) Docker Container'
	docker push $(APP_PATH)/$(APP_TAG)

.PHONY: docker-build docker-push
