package main

import (
	"authelia-basic-2fa/authelia"
	"authelia-basic-2fa/util"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Used to impersonate the client when interacting with Authelia
type ClientHandler struct {
	ctx           echo.Context
	clientCookies map[string]*http.Cookie
	proxyCookies  map[string]*http.Cookie
}

// Creates a new ClientHandler
func NewClientHandler(ctx echo.Context) *ClientHandler {
	clientCookies := map[string]*http.Cookie{}
	// save client's cookies (e.g. Authelia session) to use for sub-requests
	for _, cookie := range ctx.Cookies() {
		// ignore non-whitelisted client cookies
		if _, exists := util.CookieWhitelist[cookie.Name]; exists {
			util.SLogger.Debugf("Saving client cookie: %+v", cookie)
			clientCookies[cookie.Name] = cookie
		} else {
			util.SLogger.Debugf("NOT saving client cookie: %+v", cookie)
		}
	}
	return &ClientHandler{
		ctx:           ctx,
		clientCookies: clientCookies,
		proxyCookies:  map[string]*http.Cookie{},
	}
}

// Performs first factor authentication with Authelia and returns the JSON response status
func (a *ClientHandler) checkFirstFactor(credentials *Credentials) (bool, error) {
	return a.doStatusPost(&authelia.FirstFactorRequest{
		Username:       credentials.Username,
		Password:       credentials.Password,
		KeepMeLoggedIn: false,
	}, authelia.FirstFactorUrl, false)
}

// Performs TOTP second factor authentication with Authelia and returns the JSON response status
func (a *ClientHandler) checkTOTP(credentials *Credentials) (bool, error) {
	return a.doStatusPost(&authelia.TOTPRequest{
		Token: credentials.TOTP,
	}, authelia.TOTPUrl, false)
}

// Performs a POST request to an Authelia endpoint and returns the JSON response status
func (a *ClientHandler) doStatusPost(data interface{}, endpoint string, includeAuthorization bool) (bool, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	resp, err := a.doRequest(endpoint, "POST", jsonBody, includeAuthorization)
	if err != nil || resp.StatusCode != 200 {
		return false, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	statusResponse := authelia.StatusResponse{}
	if err = json.Unmarshal(bodyBytes, &statusResponse); err != nil {
		return false, err
	}
	if statusResponse.Status == "OK" {
		return true, nil
	}
	return false, nil
}

// Adds whitelisted headers from the client's original request to a sub-request
func (a *ClientHandler) cloneHeaders(req *http.Request, includeAuthorization bool) {
	// clone host, per
	req.Host = a.ctx.Request().Host

	// clone headers
	for key, values := range a.ctx.Request().Header {
		keyStr := strings.ToLower(key)
		if keyStr == "authorization" && !includeAuthorization {
			continue
		}

		if _, exists := util.HeaderClientWhitelist[keyStr]; exists {
			util.SLogger.Debugf("Restoring header: %s, %v", key, values)

			// Authelia expects Proxy-Authorization
			// https://github.com/authelia/authelia/blob/829757d3bc8196d6520f24479370a9037fbdb4de/internal/handlers/handler_verify.go#L232
			if keyStr == "authorization" {
				key = "Proxy-Authorization"
			}

			for _, value := range values {
				req.Header.Set(key, value)
			}
		} else {
			util.SLogger.Debugf("NOT restoring header: %s, %v", key, values)
		}
	}
}

// Saves whitelisted response cookies to ClientHandler, overwriting old ones with same name
func (a *ClientHandler) saveCookies(resp *http.Response) {
	for _, cookie := range resp.Cookies() {
		if _, exists := util.CookieWhitelist[cookie.Name]; exists {
			util.SLogger.Debugf("Saving proxy cookie: %+v", cookie)
			a.proxyCookies[cookie.Name] = cookie
		} else {
			util.SLogger.Debugf("NOT saving proxy cookie: %+v", cookie)
		}
	}
}

// Adds saved ClientHandler cookies to a request
func (a *ClientHandler) restoreCookies(req *http.Request) {
	for _, cookie := range a.clientCookies {
		// allow proxyCookies to override clientCookies
		if _, exists := a.proxyCookies[cookie.Name]; !exists {
			util.SLogger.Debugf("Restoring client cookie: %+v", cookie)
			req.AddCookie(cookie)
		} else {
			util.SLogger.Debugf("NOT restoring client cookie (proxy cookie override): %+v", cookie)
		}
	}
	for _, cookie := range a.proxyCookies {
		util.SLogger.Debugf("Restoring proxy cookie: %+v", cookie)
		req.AddCookie(cookie)
	}
}

// Performs a request to an Authelia endpoint
func (a *ClientHandler) doRequest(
	requestUri string, requestMethod string, jsonBody []byte, includeAuthorization bool) (*http.Response, error) {
	req, err := http.NewRequest(requestMethod, requestUri, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	a.cloneHeaders(req, includeAuthorization)
	a.restoreCookies(req)

	if jsonBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	a.saveCookies(resp)
	return resp, nil
}

// Checks if the client has valid Authorization
func (a *ClientHandler) checkAuthorization() (bool, map[string]string, error) {
	if a.ctx.Request().Header.Get("authorization") == "" {
		return false, nil, nil
	}
	resp, err := a.doRequest(authelia.VerifyUrl, "GET", nil, true)
	if err != nil {
		return false, nil, err
	}
	returnHeaders := a.getServerReturnHeaders(resp)
	return resp.StatusCode == 200, returnHeaders, nil
}

// Checks if the client has a valid Authelia session
func (a *ClientHandler) checkSession() (bool, map[string]string, error) {
	if _, exists := a.clientCookies[authelia.SessionCookieName]; !exists {
		if _, exists := a.proxyCookies[authelia.SessionCookieName]; !exists {
			return false, nil, nil
		}
	}
	resp, err := a.doRequest(authelia.VerifyUrl, "GET", nil, false)
	if err != nil {
		return false, nil, err
	}
	returnHeaders := a.getServerReturnHeaders(resp)
	return resp.StatusCode == 200, returnHeaders, nil
}

func (a *ClientHandler) getServerReturnHeaders(resp *http.Response) map[string]string {
	returnHeaders := map[string]string{}
	for key, values := range resp.Header {
		if _, ok := util.HeaderServerWhitelist[strings.ToLower(key)]; ok {
			returnHeaders[key] = values[0]
		}
	}
	return returnHeaders
}
