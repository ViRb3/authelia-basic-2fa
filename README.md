# Authelia Basic Auth 2FA
> Use Authelia 2-factor authentication through only standard basic auth

## Introduction
This project allows you to use [Authelia](https://github.com/authelia/authelia)'s 2FA through only [basic auth](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) and a
custom credentials format described [below](#format). This allows you to use 2FA on clients and scenarios
that demand basic auth, e.g. [webdav](https://en.wikipedia.org/wiki/WebDAV) network streaming.

## Technical details
2FA is achieved through basic auth by placing a reverse proxy (this project) before every authentication attempt with Authelia. Your requests will look like this:
```
You ---> nginx (or other reverse proxy) ---> this reverse proxy --> Authelia
```

This proxy will clone the client's request headers and cookies based on a whitelist, and use them to negotiate authentication with Authelia on the client's behalf.

The proxy will first execute a sub-request to Authelia's `verify` endpoint to check if the client has a valid session cookie or authorization (e.g. basic auth). If that succeeds, code `200` is returned to the client directly.

If that fails, the proxy will attempt to detect if the special credentials format is being used. If yes, it will decode the credentials (which include the TOTP) and execute standard Authelia 2FA TOTP authentication. The proxy will then verify the newly obtained session, and, if valid, return the session cookie to the client through a `Set-Cookie` header, along with a status code `200`.

In all other cases, including when the client does not use the special credentials format or the format is invalid, this proxy will return a status code `401`.

## Format
The custom format combines the password and TOTP into the basic auth password field. Example:

### Original credentials
- Username: `john`
- Password: `secret`
- TOTP: `123456`

### New credentials
- Username: `john`
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

## :warning: Security notes
- Make sure you are setting all reverse proxy headers from [headerWhitelist.go](util/headerWhitelist.go) in your nginx configuration, as shown in [authelia-proxy.conf](docker/nginx/data/authelia-proxy.conf). This project will pass all the headers listed above from the client to Authelia, allowing an attacker to spoof them if nginx is not present.

## Other notes
- Make sure `Set-Cookie` headers can reach the client through `auth_request` or the client will always create a new session and lose access after the TOTP expires. Check `auth_request_set` in [auth.conf](docker/nginx/data/auth.conf)
- Make sure Authelia is aware of the real client IP or you may lock out your server on bruteforce attempts. Check `set_real_ip_from` in [authelia-proxy.conf](docker/nginx/data/authelia-proxy.conf)
- Your client (e.g. [VLC Player](https://www.videolan.org/vlc/)) must support cookies and use the session cookie on subsequent requests, since the basic auth password will become invalid after the TOTP expires
