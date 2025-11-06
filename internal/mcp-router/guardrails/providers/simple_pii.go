package providers

import "regexp"

var (
	creditCardPattern = regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	awsKeyPattern     = regexp.MustCompile(`\b(AKIA[0-9A-Z]{16})\b`)
	githubKeyPattern  = regexp.MustCompile(`\b(ghp_[a-zA-Z0-9]{36}|gho_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59})\b`)
	openaiKeyPattern  = regexp.MustCompile(`\b(sk-[a-zA-Z0-9]{48})\b`)
	genericKeyPattern = regexp.MustCompile(`\b(api[_-]?key|secret[_-]?key|private[_-]?key)[\s:=]+[a-zA-Z0-9_-]{20,}\b`)
)

// SimplePIIGuardrailsProvider is a very basic Personally Identifiable
// Information (PII) detection guardrails provider. It uses simple regular
// expressions to identify potential PII in HTTP responses, such as API keys or
// other secrets.
type SimplePIIGuardrailsProvider struct{}

// NewSimplePIIGuardrailsProvider creates a new instance of
// SimplePIIGuardrailsProvider.
func NewSimplePIIGuardrailsProvider() *SimplePIIGuardrailsProvider {
	return &SimplePIIGuardrailsProvider{}
}

// ProcessResponse checks the payload for PII.
func (p *SimplePIIGuardrailsProvider) ProcessResponse(body []byte) (bool, string, error) {
	if creditCardPattern.Match(body) {
		return false, "Response contains potential credit card number", nil
	}

	if awsKeyPattern.Match(body) {
		return false, "Response contains potential AWS access key", nil
	}

	if githubKeyPattern.Match(body) {
		return false, "Response contains potential GitHub token", nil
	}

	if openaiKeyPattern.Match(body) {
		return false, "Response contains potential OpenAI API key", nil
	}

	if genericKeyPattern.Match(body) {
		return false, "Response contains potential API key or secret", nil
	}

	return true, "", nil
}
