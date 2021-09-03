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
	user := fmt.Sprint("User " + ctx.RealIP())
	util.SLogger.Debug(user + " connected")
	statusCode, returnHeaders, err := checkAuthentication(ctx)
	if err != nil {
		util.SLogger.Error(user + " not authenticated")
		util.SLogger.Error(err)
		return ctx.NoContent(500)
	}
	if util.IsGood(statusCode) {
		util.SLogger.Info(user + " authenticated")
		for key, value := range returnHeaders {
			ctx.Response().Header().Set(key, value)
		}
		return ctx.NoContent(statusCode)
	} else {
		util.SLogger.Info(user + " not authenticated")
		return ctx.NoContent(statusCode)
	}
}

func checkAuthentication(ctx echo.Context) (int, map[string]string, error) {
	clientHandler := NewClientHandler(ctx)
	// apply all proxyCookies to the response, e.g. newly created Authelia session
	defer func() {
		for _, cookie := range clientHandler.proxyCookies {
			util.SLogger.Debugf("Applying proxy cookie: %+v", cookie)
			ctx.SetCookie(cookie)
		}
	}()

	util.SLogger.Debug("Checking if user session is already valid")
	statusCode, returnHeaders, err := clientHandler.checkSession()
	if err != nil {
		return 0, nil, err
	}
	if util.IsGood(statusCode) {
		util.SLogger.Debug("User session was valid")
		return statusCode, returnHeaders, nil
	} else if statusCode != 401 {
		return statusCode, nil, nil
	}

	util.SLogger.Debug("Checking if user authorization is valid")
	statusCode, returnHeaders, err = clientHandler.checkAuthorization()
	if err != nil {
		return 0, nil, err
	}
	if util.IsGood(statusCode) {
		util.SLogger.Debug("Authorization was valid")
		return statusCode, returnHeaders, nil
	} else if statusCode != 401 {
		return statusCode, nil, nil
	}

	util.SLogger.Debug("Performing manual authentication")
	credentials, err := DecodeCredentials(ctx)
	if err != nil {
		util.SLogger.Debug(err)
		return 401, nil, nil
	}
	util.SLogger.Debug("Checking first factor authentication")
	statusCode, err = clientHandler.checkFirstFactor(credentials)
	if err != nil {
		return 0, nil, err
	}
	if util.IsBad(statusCode) {
		return statusCode, nil, nil
	}
	util.SLogger.Debug("Checking TOTP authentication")
	statusCode, err = clientHandler.checkTOTP(credentials)
	if err != nil {
		return 0, nil, err
	}
	if util.IsBad(statusCode) {
		return statusCode, nil, nil
	}

	util.SLogger.Debug("Checking if new session is valid")
	return clientHandler.checkSession()
}
