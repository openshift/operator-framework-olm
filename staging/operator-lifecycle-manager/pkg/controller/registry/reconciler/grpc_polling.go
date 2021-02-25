package reconciler

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/queueinformer"
)

// SyncRegistryUpdateInterval returns a duration to use when requeuing the catalog source for reconciliation.
// This ensures that the catalog is being synced on the correct time interval based on its spec.
// Note: this function assumes the catalog has an update strategy set.
func SyncRegistryUpdateInterval(source *v1alpha1.CatalogSource, now time.Time) time.Duration {
	pollingInterval := source.Spec.UpdateStrategy.Interval.Duration
	latestPoll := source.Status.LatestImageRegistryPoll
	creationTimestamp := source.CreationTimestamp.Time

	// Resync before next default sync if the polling interval is less than the default
	if pollingInterval <= queueinformer.DefaultResyncPeriod {
		return pollingInterval
	}
	// Resync based on the delta between the default sync and the actual last poll if the interval is greater than the default
	return defaultOr(latestPoll, pollingInterval, creationTimestamp, now)
}

// defaultOr returns either the default resync period or the time remaining until the next poll is due, whichever is smaller.
// For example, if the polling interval is 40 minutes, OLM will sync after ~30 minutes and see that the next sync
// for this catalog should be in 10 minutes, sooner than the default 15 minutes, and return 10 minutes.
func defaultOr(latestPoll *metav1.Time, pollingInterval time.Duration, creationTimestamp time.Time, now time.Time) time.Duration {
	if latestPoll.IsZero() {
		latestPoll = &metav1.Time{Time: creationTimestamp}
	}

	remaining := latestPoll.Add(pollingInterval).Sub(now)
	// sync ahead of the default interval in the case where now + default sync is after the last sync plus the interval
	if remaining < queueinformer.DefaultResyncPeriod {
		return remaining
	}
	// return the default sync period otherwise: the next sync cycle will check again
	return queueinformer.DefaultResyncPeriod
}
