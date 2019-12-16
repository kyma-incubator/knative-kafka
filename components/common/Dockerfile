FROM golang:1.13.5-buster as builder

# Gather the librdkafka and knative-kafka-dispatcher dependencies for building
RUN wget -qO - https://packages.confluent.io/deb/5.3/archive.key | apt-key add - \
    && echo "deb https://packages.confluent.io/deb/5.3 stable main" >> /etc/apt/sources.list \
    && echo "deb http://security.debian.org/debian-security jessie/updates main" >> /etc/apt/sources.list \
    && apt-get update \
    && apt-get -y install librdkafka-dev

# Copy Over Only This Component
WORKDIR /go/src/github.com/kyma-incubator/knative-kafka/components/common
ADD components/common .

ENV GOPATH=/go

# Perform The Build Via The Makefile
RUN /bin/bash -c "cd /go/src/github.com/kyma-incubator/knative-kafka/components/common && make test"
