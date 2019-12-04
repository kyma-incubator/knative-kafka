/*
.
*/

package controller

import (
	kafkaadmin "github.com/kyma-incubator/knative-kafka/components/common/pkg/kafka/admin"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, kafkaadmin.AdminClientInterface) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, adminClient kafkaadmin.AdminClientInterface) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, adminClient); err != nil {
			return err
		}
	}
	return nil
}
