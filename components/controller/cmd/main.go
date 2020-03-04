package main

import (
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/controller/kafkachannel"
	"knative.dev/pkg/injection/sharedmain"
)

func main() {
	// ensure admin clients are shut down gracefully
	defer kafkachannel.Shutdown()

	sharedmain.Main("knativekafkachannel_controller", kafkachannel.NewController)
}
