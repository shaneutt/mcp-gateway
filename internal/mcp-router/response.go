package mcprouter

import (
	"fmt"
	"log/slog"
	"net/http"

	basepb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	eppb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	typepb "github.com/envoyproxy/go-control-plane/envoy/type/v3"
)

// passthroughResponse is an empty response to signal to continue processing.
var passthroughResponse = &eppb.ProcessingResponse{
	Response: &eppb.ProcessingResponse_ResponseBody{
		ResponseBody: &eppb.BodyResponse{},
	},
}

// extractHelperSessionFromBackend extracts the helper session ID from a backend session ID
// Returns empty string if not a backend session ID
func extractHelperSessionFromBackend(_ string) string {
	// TODO: check known server session prefixes
	return ""
}

// HandleResponseHeaders handles response headers for session ID reverse mapping
func (s *ExtProcServer) HandleResponseHeaders(
	headers *eppb.HttpHeaders) ([]*eppb.ProcessingResponse, error) {
	slog.Info("[EXT-PROC] Processing response headers for session mapping...", "headers", headers)

	if headers == nil || headers.Headers == nil {
		slog.Info("[EXT-PROC] No response headers to process")
		return []*eppb.ProcessingResponse{
			{
				Response: &eppb.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: &eppb.HeadersResponse{},
				},
			},
		}, nil
	}

	// Look for mcp-session-id header that needs reverse mapping
	mcpSessionID := getSingleValueHeader(headers.Headers, "mcp-session-id")

	if mcpSessionID == "" {
		slog.Info("[EXT-PROC] No mcp-session-id in response headers")
		return []*eppb.ProcessingResponse{
			{
				Response: &eppb.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: &eppb.HeadersResponse{},
				},
			},
		}, nil
	}

	slog.Info(fmt.Sprintf("[EXT-PROC] Response backend session: %s", mcpSessionID))

	// Check if this is a backend session that needs mapping back to helper session
	helperSession := extractHelperSessionFromBackend(mcpSessionID)
	if helperSession == "" {
		// Not a backend session ID, leave as-is
		slog.Info("[EXT-PROC] Session ID doesn't need reverse mapping")
		return []*eppb.ProcessingResponse{
			{
				Response: &eppb.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: &eppb.HeadersResponse{},
				},
			},
		}, nil
	}

	slog.Info(
		fmt.Sprintf("[EXT-PROC] Mapping backend session back to helper session: %s", helperSession),
	)

	// Return response with updated session header
	return []*eppb.ProcessingResponse{
		{
			Response: &eppb.ProcessingResponse_ResponseHeaders{
				ResponseHeaders: &eppb.HeadersResponse{
					Response: &eppb.CommonResponse{
						HeaderMutation: &eppb.HeaderMutation{
							SetHeaders: []*basepb.HeaderValueOption{
								{
									Header: &basepb.HeaderValue{
										Key:      "mcp-session-id",
										RawValue: []byte(helperSession),
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

// -----------------------------------------------------------------------------
// Response Handling - Guardrails Response Processor
// -----------------------------------------------------------------------------

// ProcessGuardrailsForResponse processes the response with the configured
// guardrail providers.
func (s *ExtProcServer) ProcessGuardrailsForResponse(
	httpBody *eppb.HttpBody) ([]*eppb.ProcessingResponse, error) {
	body := httpBody.GetBody()

	slog.Info(fmt.Sprintf("[EXT-PROC] Processing %d guardrails providers for response (size: %d, end_of_stream: %t)",
		len(s.GuardrailsProviders), len(body), httpBody.GetEndOfStream()))

	for providerName, provider := range s.GuardrailsProviders {
		slog.Info(fmt.Sprintf("[EXT-PROC] Processing guardrails provider %s for response", providerName))

		flagged, details, err := provider.ProcessResponse(body)
		if err != nil {
			slog.Error("[EXT-PROC] Guardrails provider error", "error", err, "provider", providerName)
			return extProcErrorResponse(providerName, err), nil
		}

		if flagged {
			slog.Info("[EXT-PROC] Guardrails provider flagged the response")
			return extProcFlaggedResponse(providerName, details), nil
		}
	}

	slog.Info("[EXT-PROC] Response accepted by guardrails providers")
	return []*eppb.ProcessingResponse{passthroughResponse}, nil
}

// -----------------------------------------------------------------------------
// Response Handling - Guardrails ExtProc Utilities
// -----------------------------------------------------------------------------

// extProcErrorResponse constructs an ext_proc error response for guardrails
// provider errors.
func extProcErrorResponse(providerName string, err error) []*eppb.ProcessingResponse {
	details := fmt.Sprintf("guardrails provider %s error: %s", providerName, err)
	return []*eppb.ProcessingResponse{
		{
			Response: &eppb.ProcessingResponse_ImmediateResponse{
				ImmediateResponse: &eppb.ImmediateResponse{
					Status: &typepb.HttpStatus{
						Code: typepb.StatusCode(http.StatusServiceUnavailable),
					},
					Body:    []byte(details),
					Details: details,
				}}}}
}

// extProcFlaggedResponse constructs an ext_proc response for guardrails for
// flagged responses.
func extProcFlaggedResponse(providerName, details string) []*eppb.ProcessingResponse {
	details = fmt.Sprintf("Response flagged by guardrails provider %s: %s", providerName, details)
	return []*eppb.ProcessingResponse{
		{
			Response: &eppb.ProcessingResponse_ImmediateResponse{
				ImmediateResponse: &eppb.ImmediateResponse{
					Status: &typepb.HttpStatus{
						Code: typepb.StatusCode(http.StatusUnavailableForLegalReasons),
					},
					Body:    []byte(details),
					Details: details,
				}}}}
}
