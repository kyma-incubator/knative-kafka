FROM ubuntu:18.04 as builder

# Gather the librdkafka and konduit-kafka-common dependencies for building
RUN apt-get update \
    && apt-get -y install software-properties-common \
    && add-apt-repository ppa:longsleep/golang-backports \
    && apt-get -y install curl \
    golang-go \
    libsasl2-dev \
    libssl-dev \
    libzstd-dev \
    openssl \
    python-dev \
    zlib1g-dev

# Build librdkafka using static libraries when possible to reduce image size
WORKDIR /librdkafka-tmp
RUN curl -LO https://github.com/edenhill/librdkafka/archive/v1.0.0.tar.gz \
    && tar -xzf v1.0.0.tar.gz \
    && mv librdkafka-1.0.0 librdkafka \
    && cd librdkafka \
    && STATIC_LIB_zstd=/usr/lib/x86_64-linux-gnu/libzstd.a ./configure --enable-static \
    && make \
    && make install

# Copy Over Only This Component
WORKDIR /go/src/github.com/kyma-incubator/knative-kafka/components/common
ADD components/common .

ENV GOPATH=/go

# Perform The Build Via The Makefile
RUN /bin/bash -c "cd /go/src/github.com/kyma-incubator/knative-kafka/components/common && make test"
