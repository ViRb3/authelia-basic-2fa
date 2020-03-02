# Authelia Basic Auth 2FA
> Use Authelia 2FA through only standard basic auth

## Introduction
This project allows you to use [Authelia](https://github.com/authelia/authelia)'s 2FA through only [basic auth](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) and a
custom credentials format described [below](#format). This allows you to use 2FA on clients and scenarios
that demand basic auth, e.g. [webdav](https://en.wikipedia.org/wiki/WebDAV) network streaming.

## Changes from v1
This is a complete re-code of the original LUA script into a Go reverse proxy, fixing numerous compatibility and security issues. You are _strongly_ urged to upgrade. As an architectural bonus, you no longer need OpenResty or even nginx to use this project.

## Technical details
2FA is achieved through basic auth by placing a reverse proxy (this project) before every authentication attempt with Authelia. Your requests will look like this:
```
You ---> nginx (or other reverse proxy) ---> this reverse proxy --> Authelia
```

If the client has provided an Authelia session cookie, this proxy will first execute a sub-request to Authelia's `verify` endpoint to validate the session. If that succeeds, code `200` is returned directly.

If the session is invalid or no such exists, this proxy will attempt to detect if the special credentials format is being used. If yes, it will decode them and execute standard Authelia 2FA authentication on behalf of the client using sub-requests. this proxy will finally return the session cookie to the client through a `Set-Cookie` header, along with a status code `200`.

In all other cases, including when the client does not use the special credentials format or the format is invalid, this proxy will return a status code `401`.

## Format
The custom format combines the password and TOTP into the basic auth password field. To hint the backend you are attempting this 'special' form of authentication, you suffix your password with an underscore ( _ ). This can be changed in the [source code](credentials.go).

### Original credentials
- Username: `john`
- Password: `secret`
- TOTP: `123456`

### New credentials
- Username: `john_`
- Password: `secret123456`

## Requirements
- [Nginx](https://www.nginx.com/) (or any other reverse proxy)
- [Authelia](https://github.com/authelia/authelia)

## Installation
Check out the [Docker guide](docker). If you do not use Docker, you can still extract the configuration and use it directly.

## Usage
Run with argument `-help`:
```bash
-debug
    Debug logging
-ip string
    Listening ip (default "0.0.0.0")
-port int
    Listening port (default 8081)
-url string
    Authelia URL to use for authentication (default "http://authelia:9091")
```

## Notes
- Make sure `Set-Cookie` headers can reach the client through `auth_request` or the client will always create a new session and lose access after the TOTP expires. Check `auth_request_set` in [auth.conf](auth.conf)
- Make sure Authelia is aware of the real client IP or you may lock out your server on bruteforce attempts. Check `set_real_ip_from` in [auth.conf](auth.conf)
- Your client (e.g. [VLC Player](https://www.videolan.org/vlc/)) must support cookies and use the session cookie on subsequent requests, since the basic auth password will become invalid after the TOTP expires
