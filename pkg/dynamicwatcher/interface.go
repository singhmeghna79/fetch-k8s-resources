package dynamicwatcher

import (
	"time"
)

// WatcherInterface contains methods that maintains watcher life cycle
type WatcherInterface interface {
	// Run maintains full lifecycle of a dynamic watcher. If any error occurs it
	// calls schedule run and tries to re-run . For now it waits constant time and
	// do a re run. We can put more intelligence with wait time later. Main intention
	// to have it - you call run and it will do self heal for any error.
	Run()
	// ScheduleRun calls run after given duration.
	ScheduleRun(d time.Duration)
	// Verify checks if we can start watcher for a given details.
	Verify() error
	// Watch contains start and stop watcher functionality.
	Watch(stopCh <-chan struct{})
	// Stop is the only way to stop a watcher. If you called stop it will not be
	// re launched automatically.
	Stop()
}
