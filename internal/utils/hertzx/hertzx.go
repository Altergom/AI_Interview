package hertzx

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

const bearerPrefix = "Bearer "

// ClientIP returns the real client IP from Hertz request context.
// It prefers X-Forwarded-For / X-Real-IP because services are commonly deployed behind reverse proxies.
func ClientIP(c *app.RequestContext) string {
	if xff := string(c.GetHeader("X-Forwarded-For")); xff != "" {
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := string(c.GetHeader("X-Real-IP")); xri != "" {
		return strings.TrimSpace(xri)
	}
	return c.ClientIP()
}

// BearerFromAuthorization extracts bearer token from Authorization header value.
// It returns empty string when header is missing or malformed.
func BearerFromAuthorization(header string) string {
	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if token == "" {
		return ""
	}
	return token
}

// BearerToken extracts bearer token from request Authorization header.
func BearerToken(c *app.RequestContext) string {
	return BearerFromAuthorization(string(c.GetHeader("Authorization")))
}

