package consensus

import "github.com/prysmaticlabs/prysm/shared/event"

// Notifier interface defines the methods of the service that provides minimal consensus info updates to consumers.
type Notifier interface {
	ConsensusFeed() *event.Feed
}
