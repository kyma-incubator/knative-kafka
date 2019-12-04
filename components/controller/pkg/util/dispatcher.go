package util

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"regexp"
	"strings"
)

// Compile RegExps
var startsWithLowercaseAlphaCharRegExp = regexp.MustCompile("^[a-z].*$")
var endsWithLowercaseAlphaCharRegExp = regexp.MustCompile("^.*[a-z]$")
var invalidK8sServiceCharactersRegExp = regexp.MustCompile("[^a-z0-9\\-]+")

// Dispatcher K8S Service Naming Utility
func DispatcherServiceName(subscription *eventingv1alpha1.Subscription) string {
	return generateValidDispatcherDnsName(subscription.Name)
}

// Dispatcher Deployment Naming Utility
func DispatcherDeploymentName(subscription *eventingv1alpha1.Subscription) string {
	return generateValidDispatcherDnsName(subscription.Name)
}

// Return A Valid Dispatcher DNS Name Which Is As Close To The Specified Name As Possible
func generateValidDispatcherDnsName(name string) string {

	// Convert To LowerCase
	validDnsName := strings.ToLower(name)

	// Strip Any Invalid DNS Characters
	validDnsName = invalidK8sServiceCharactersRegExp.ReplaceAllString(validDnsName, "")

	// Prepend Alpha Prefix If Needed
	if !startsWithLowercaseAlphaCharRegExp.MatchString(validDnsName) {
		validDnsName = "kk-" + validDnsName // Should never happen but just in case ; )
	}

	// Append The Dispatcher Suffix
	validDnsName = validDnsName + "-dispatcher"

	// Truncate If Too Long
	if len(validDnsName) > 55 {
		validDnsName = validDnsName[:55]
	}

	// Remove Any Trailing Non Alpha
	for !endsWithLowercaseAlphaCharRegExp.MatchString(validDnsName) {
		validDnsName = validDnsName[:(len(validDnsName) - 1)]
	}

	// Return The Valid DNS Name
	return validDnsName
}
