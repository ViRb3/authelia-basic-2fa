package main

import (
	"authelia-basic-2fa/authelia"
	"flag"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
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
		e.Logger.SetLevel(log.DEBUG)
	} else {
		e.Logger.SetLevel(log.INFO)
	}
	e.GET("*", handleAuthentication)

	e.Logger.Info("Using Authelia URL: ", *url)
	e.Logger.Info("Listening on: ", listenAddress)
	e.Logger.Fatal(e.Start(listenAddress))
}

func handleAuthentication(ctx echo.Context) error {
	ctx.Logger().Debug("User connected")
	authenticated, err := checkAuthentication(ctx)
	if err != nil {
		ctx.Logger().Error("User not authenticated")
		ctx.Logger().Error(err)
		return ctx.NoContent(401)
	}
	if authenticated {
		ctx.Logger().Info("User authenticated")
		return ctx.NoContent(200)
	}
	ctx.Logger().Info("User not authenticated")
	return ctx.NoContent(401)
}

func checkAuthentication(ctx echo.Context) (bool, error) {
	clientHandler := NewClientHandler(ctx)
	// apply all proxyCookies to the response, e.g. newly created Authelia session
	defer func() {
		for _, cookie := range clientHandler.proxyCookies {
			ctx.Logger().Debugf("Applying proxy cookie: %+v", cookie)
			ctx.SetCookie(cookie)
		}
	}()

	ctx.Logger().Debug("Checking if user session is already valid")
	sessionValid, err := clientHandler.checkSession()
	if err != nil {
		return false, err
	}
	if sessionValid {
		ctx.Logger().Debug("User session was valid")
		return true, nil
	}

	ctx.Logger().Debug("Checking if user authorization is valid")
	authorizationValid, err := clientHandler.checkAuthorization()
	if err != nil {
		return false, err
	}
	if authorizationValid {
		ctx.Logger().Debug("Authorization was valid")
		return true, nil
	}

	ctx.Logger().Debug("Performing manual authentication")
	credentials, err := DecodeCredentials(ctx)
	if err != nil {
		return false, err
	}
	ctx.Logger().Debug("Checking first factor authentication")
	result, err := clientHandler.checkFirstFactor(credentials)
	if err != nil || !result {
		return false, err
	}
	ctx.Logger().Debug("Checking TOTP authentication")
	result, err = clientHandler.checkTOTP(credentials)
	if err != nil || !result {
		return false, err
	}

	ctx.Logger().Debug("Checking if new session is valid")
	return clientHandler.checkSession()
}
