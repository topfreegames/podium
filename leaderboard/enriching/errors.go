package enriching

import "errors"

var (
	// ErrEnrichmentCall is returned when the webhook call fails on the side of the webhook.
	ErrEnrichmentCall = errors.New("the call to the webhook returned an error response")

	// ErrEnrichmentInternal is returned when the enrichment fails for an internal reason.
	ErrEnrichmentInternal = errors.New("could not perform enrichment")
)
