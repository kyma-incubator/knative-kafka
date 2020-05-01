# Makefile For Prow Compatibility Purposes, Just Defers To Individual Components Makefile

BUILD_ROOT:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

#
# Docker Targets
#

docker-build-all:
	@echo 'Building Channel Docker Image'
	cd $(BUILD_ROOT)/components/channel; make docker-build
	@echo 'Building Dispatcher Docker Image'
	cd $(BUILD_ROOT)/components/dispatcher; make docker-build
	@echo 'Building Controller Docker Image'
	cd $(BUILD_ROOT)/components/controller; make docker-build

docker-push-all:
	@echo 'Pushing Channel Docker Image'
	cd $(BUILD_ROOT)/components/channel; make docker-push
	@echo 'Pushing Dispatcher Docker Image'
	cd $(BUILD_ROOT)/components/dispatcher; make docker-push
	@echo 'Pushing Controller Docker Image'
	cd $(BUILD_ROOT)/components/controller; make docker-push

.PHONY: docker-build-all docker-push-all


#
# Support Prow Targets
#

ci-master: docker-push-all

ci-release: docker-push-all

ci-pr: docker-build-all

.PHONY: ci-master ci-release ci-pr
