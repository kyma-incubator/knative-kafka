package testing

import (
	"github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/tools/record"

	fakeclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned/fake"
	"knative.dev/pkg/controller"
	. "knative.dev/pkg/reconciler/testing"
)

const (
	// maxEventBufferSize is the estimated max number of event notifications that
	// can be buffered during reconciliation.
	maxEventBufferSize = 10
)

// Ctor functions create a k8s controller with given params.
type Ctor func(listers *Listers, kafkaClient versioned.Interface, eventRecorder record.EventRecorder) controller.Reconciler

// MakeFactory creates a reconciler factory with fake clients and controller created by `ctor`.
func MakeFactory(ctor Ctor) Factory {
	return func(t *testing.T, r *TableRow) (controller.Reconciler, ActionRecorderList, EventList, *FakeStatsReporter) {
		ls := NewListers(r.Objects)

		client := fakeclientset.NewSimpleClientset(ls.GetMessagingObjects()...)

		dynamicScheme := runtime.NewScheme()
		for _, addTo := range clientSetSchemes {
			addTo(dynamicScheme)
		}

		eventRecorder := record.NewFakeRecorder(maxEventBufferSize)
		statsReporter := &FakeStatsReporter{}

		// Set up our Controller from the fakes.
		c := ctor(&ls, client, eventRecorder)

		actionRecorderList := ActionRecorderList{client}
		eventList := EventList{Recorder: eventRecorder}

		return c, actionRecorderList, eventList, statsReporter
	}
}
