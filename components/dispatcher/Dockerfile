FROM ubuntu:18.04 as builder

# Gather the librdkafka and knative-kafka-dispatcher dependencies for building
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
WORKDIR /go/src/github.com/kyma-incubator/knative-kafka/components/dispatcher
ADD components/dispatcher .

# Get The RUN_TESTS Argument
ARG RUN_TESTS=false

# Perform The Build Via The Makefile (Optionally Run Tests)
RUN /bin/bash -c "cd /go/src/github.com/kyma-incubator/knative-kafka/components/dispatcher && \
source ./local-env.sh && \
if [[ \"${RUN_TESTS}\" == \"true\" ]]; then \
  make test build-native; \
else \
  make build-native; \
fi"

# Determine Dependencies of knative-kafka-dispatcher that will be needed for the final image and package them into one dir
WORKDIR /deps
RUN ldd /go/src/github.com/kyma-incubator/knative-kafka/components/dispatcher/build/knative-kafka-dispatcher \
    | tr -s '[:blank:]' '\n' \
    | grep '^/' \
    | xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'

# Other dependencies that aren't found using the ldd command. These could be transitive dependencies.
# These were found at runtime generally, when the image reported runtime errors.
RUN cp /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 ./lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 \
    && cp /lib/x86_64-linux-gnu/libnss_dns.so.2 ./lib/x86_64-linux-gnu/libnss_dns.so.2 \
    && cp /lib/x86_64-linux-gnu/libresolv.so.2 ./lib/x86_64-linux-gnu/libresolv.so.2 \
    && cp /usr/lib/x86_64-linux-gnu/libsasl2.so.2 ./usr/lib/x86_64-linux-gnu/libsasl2.so.2
RUN mkdir -p lib64 \
    && cp /lib64/ld-linux-x86-64.so.2 lib64/ld-linux-x86-64.so.2


# Create Docker Container From Google's distroless base
FROM gcr.io/distroless/base

# Manage malloc and os thread count that can execute go code.
ENV MALLOC_ARENA_MAX=1 GOMAXPROCS=1

# Copy over the knative-kafka binary
COPY --from=builder /go/src/github.com/kyma-incubator/knative-kafka/components/dispatcher/build/knative-kafka-dispatcher /

# Copy over the dependencies
COPY --from=builder /deps/ /

CMD [ "/knative-kafka-dispatcher" ]
