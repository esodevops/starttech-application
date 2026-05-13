package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCookieDomain determines the appropriate cookie domain based on the request host
// and the list of allowed domains.
func GetCookieDomain(c *gin.Context, allowedDomains []string) string {
	host := c.Request.Host
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	for _, domain := range allowedDomains {
		if domain == host {
			return domain
		}
		// Also allow subdomains if the allowed domain starts with a dot (optional, but good practice)
		// or if we want to be strict, just exact match.
		// For now, let's do exact match as per "list of allowed hosts".
	}

	// Fallback: use a host-only cookie bound to the current request host.
	// Returning an empty domain tells browsers to scope the cookie to the current host,
	// which is safer than forcing an unrelated value like "localhost" in production.
	return ""
}
