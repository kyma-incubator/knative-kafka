# ###############################################################################
#
#                       Knative-Kafka Common Makefile
#
# ###############################################################################

# Project Directories
BUILD_ROOT:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR:=$(BUILD_ROOT)/build

# Application Names (Align With Knative Eventing Version)
APP_NAME=knative-kafka-common
APP_VERSION=v0.13.5

# Build / Packaging / Deployment Configuration
APP_TAG=$(APP_NAME):$(APP_VERSION)

#
# Dep / Fmt / Vet / Lint / Test / Sonar
#

dep:
	@echo 'Ensuring Dependencies'
	cd $(BUILD_ROOT); dep ensure -v

fmt:
	@echo 'Formatting Go Source'
	go fmt ./...

vet:
	@echo 'Vetting Packages'
	go vet ./pkg/...

lint:
	@echo 'Linting Packages'
	golint ./pkg/...

test:
	@echo 'Testing Packages'
	cd $(BUILD_ROOT); mkdir -p $(BUILD_DIR);
	go test -v ./pkg/... -coverprofile ${BUILD_DIR}/coverage.out
	go tool cover -func=${BUILD_DIR}/coverage.out

.PHONY: dep fmt vet lint test

#
# Build Docker Container (Testing Only - No Delivery)
#

docker-build:
	@echo 'Building $(APP_NAME) Docker Container'
	docker build --rm -t $(APP_TAG) -f ./Dockerfile ../..

.PHONY: docker-build

#
# Support Prow
#

ci-master: docker-build

ci-pr: docker-build

.PHONY: ci-release ci-pr
