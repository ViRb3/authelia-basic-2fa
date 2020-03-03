package util

import "strings"

// Headers that should be passed from the client to sub-requests to Authelia
var HeaderWhitelist = map[string]bool{}

func init() {
	// taken from official nginx guide:
	// https://github.com/authelia/authelia/blob/master/docs/deployment/supported-proxies/nginx.md
	headers := []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"X-Forwarded-Proto",
		"X-Forwarded-Host",
		"X-Forwarded-Uri",
		"X-Forwarded-Ssl",
		"Connection",
		// allow e.g. basic auth
		"Authorization",
	}
	for _, header := range headers {
		HeaderWhitelist[strings.ToLower(header)] = true
	}
}
