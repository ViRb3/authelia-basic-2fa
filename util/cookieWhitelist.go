package util

// Cookies that should be passed to sub-requests to Authelia
var CookieWhitelist = map[string]bool{}

func init() {
	cookies := []string{
		"authelia_session",
	}
	for _, cookie := range cookies {
		CookieWhitelist[cookie] = true
	}
}
