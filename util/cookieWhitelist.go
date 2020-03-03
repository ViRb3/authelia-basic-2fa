package util

import "authelia-basic-2fa/authelia"

// Cookies that should be passed to sub-requests to Authelia
var CookieWhitelist = map[string]bool{}

func init() {
	cookies := []string{
		authelia.SessionCookieName,
	}
	for _, cookie := range cookies {
		CookieWhitelist[cookie] = true
	}
}
