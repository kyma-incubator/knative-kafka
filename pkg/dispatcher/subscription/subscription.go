package subscription

import eventingduck "knative.dev/eventing/pkg/apis/duck/v1alpha1"

// TODO - breaking this out into it's own package is... ok but maybe should look at restructuring/combining the dispatcher and client?

type Subscription struct {
	eventingduck.SubscriberSpec
	GroupId string
}
