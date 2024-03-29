package main

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/labstack/echo/v4"
)

// Decodes credentials from the client's request using the custom format
func DecodeCredentials(ctx echo.Context) (*Credentials, error) {
	authHeader := ctx.Request().Header.Get("authorization")
	authHeaderSplit := strings.Split(authHeader, " ")
	if len(authHeaderSplit) != 2 {
		return nil, errors.New("unrecognized auth header format")
	}

	if strings.ToLower(authHeaderSplit[0]) != "basic" {
		return nil, errors.New("not auth basic")
	}

	authDecoded, err := base64.StdEncoding.DecodeString(authHeaderSplit[1])
	if err != nil {
		return nil, err
	}

	authString := string(authDecoded)
	authSplit := strings.Split(authString, ":")
	if len(authSplit) != 2 {
		return nil, errors.New("unrecognized auth header content format")
	}

	username := authSplit[0]
	password := authSplit[1]

	if len(password) < 7 {
		return nil, errors.New("unrecognized pass format or not special auth")
	}

	totp := password[len(password)-6:]
	password = password[:len(password)-6]

	return &Credentials{
		Username: username,
		Password: password,
		TOTP:     totp,
	}, nil
}
