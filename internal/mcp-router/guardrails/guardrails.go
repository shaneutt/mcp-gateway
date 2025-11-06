package guardrails

// Provider is the inferface for pluggable guardrails providers.
type Provider interface {
	// ProcessResponse processes the response body and indicates whether that
	// body is flagged by the provider. If the body is flagged it will also
	// provide details as to the reason.
	ProcessResponse(body []byte) (flagged bool, details string, err error)
}
