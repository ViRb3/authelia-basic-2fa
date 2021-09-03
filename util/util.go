package util

import (
	"authelia-basic-2fa/authelia"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

var Logger *zap.Logger
var SLogger *zap.SugaredLogger

func InitializeLogger(logLevel zapcore.Level) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	Logger, _ = config.Build()
	SLogger = Logger.Sugar()
}

// Cookies that should be passed to sub-requests to Authelia
var CookieWhitelist = map[string]bool{}

// Headers that should be passed from the client to sub-requests to Authelia
var HeaderClientWhitelist = map[string]bool{}

// Headers that should be passed from the server to the client who sent the request
var HeaderServerWhitelist = map[string]bool{}

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
		HeaderClientWhitelist[strings.ToLower(header)] = true
	}
}

func init() {
	cookies := []string{
		authelia.SessionCookieName,
	}
	for _, cookie := range cookies {
		CookieWhitelist[cookie] = true
	}
}

func init() {
	// taken from official nginx guide:
	// https://github.com/authelia/authelia/blob/master/docs/deployment/supported-proxies/nginx.md
	headers := []string{
		"Remote-User",
		"Remote-Groups",
		"Remote-Name",
		"Remote-Email",
	}
	for _, header := range headers {
		HeaderServerWhitelist[strings.ToLower(header)] = true
	}
}
