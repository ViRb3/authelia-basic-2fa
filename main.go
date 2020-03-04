package main

import (
	"authelia-basic-2fa/authelia"
	"authelia-basic-2fa/util"
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap/zapcore"
)

func main() {
	url := flag.String("url", "http://authelia:9091", "Authelia URL to use for authentication")
	port := flag.Int("port", 8081, "Listening port")
	ip := flag.String("ip", "0.0.0.0", "Listening ip")
	debug := flag.Bool("debug", false, "Debug logging")
	flag.Parse()

	authelia.BuildUrls(*url)
	listenAddress := fmt.Sprintf("%s:%d", *ip, *port)

	e := echo.New()
	e.HideBanner = true
	if *debug {
		util.InitializeLogger(zapcore.DebugLevel)
	} else {
		util.InitializeLogger(zapcore.InfoLevel)
	}
	e.GET("*", handleAuthentication)

	util.SLogger.Info("Using Authelia URL: " + *url)
	util.SLogger.Info("Listening on: " + listenAddress)
	util.SLogger.Fatal(e.Start(listenAddress))
}

func handleAuthentication(ctx echo.Context) error {
	util.SLogger.Debug("User connected")
	authenticated, err := checkAuthentication(ctx)
	if err != nil {
		util.SLogger.Error("User not authenticated")
		util.SLogger.Error(err)
		return ctx.NoContent(401)
	}
	if authenticated {
		util.SLogger.Info("User authenticated")
		return ctx.NoContent(200)
	}
	util.SLogger.Info("User not authenticated")
	return ctx.NoContent(401)
}

func checkAuthentication(ctx echo.Context) (bool, error) {
	clientHandler := NewClientHandler(ctx)
	// apply all proxyCookies to the response, e.g. newly created Authelia session
	defer func() {
		for _, cookie := range clientHandler.proxyCookies {
			util.SLogger.Debugf("Applying proxy cookie: %+v", cookie)
			ctx.SetCookie(cookie)
		}
	}()

	util.SLogger.Debug("Checking if user session is already valid")
	sessionValid, err := clientHandler.checkSession()
	if err != nil {
		return false, err
	}
	if sessionValid {
		util.SLogger.Debug("User session was valid")
		return true, nil
	}

	util.SLogger.Debug("Checking if user authorization is valid")
	authorizationValid, err := clientHandler.checkAuthorization()
	if err != nil {
		return false, err
	}
	if authorizationValid {
		util.SLogger.Debug("Authorization was valid")
		return true, nil
	}

	util.SLogger.Debug("Performing manual authentication")
	credentials, err := DecodeCredentials(ctx)
	if err != nil {
		return false, err
	}
	util.SLogger.Debug("Checking first factor authentication")
	result, err := clientHandler.checkFirstFactor(credentials)
	if err != nil || !result {
		return false, err
	}
	util.SLogger.Debug("Checking TOTP authentication")
	result, err = clientHandler.checkTOTP(credentials)
	if err != nil || !result {
		return false, err
	}

	util.SLogger.Debug("Checking if new session is valid")
	return clientHandler.checkSession()
}
