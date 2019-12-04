package controller

import (
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/controller/kafkachannel"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kafkachannel.Add)
}
