# Authelia Basic Auth 2FA
> Use Authelia 2FA through only standard basic auth

## Description
This project allows you to use [Authelia](https://github.com/authelia/authelia)'s 2FA through only [basic auth](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) and a
custom credentials format described [below](#format). This allows you to use 2FA on clients and scenarios
that demand basic auth, e.g. [webdav](https://en.wikipedia.org/wiki/WebDAV) network streaming.

## Technical details
The way 2FA is achieved through basic auth is a [LUA](https://www.lua.org/about.html) script that runs before every authentication attempt with Authelia. The script will attempt to detect if the special credentials format is being used. If yes, it will decode them and execute standard Authelia 2FA authentication on behalf of the client using nginx subrequests. It will finally return the session cookie to the client, along with a status code `200` to go through the `auth_request`. If the client does not use the special credentials format or the format is invalid, the script will return and pass the whole request to the standard Authelia `verify` endpoint. 

## Format
The custom format combines the password and TOTP into the basic auth password field. To hint the backend you are attempting this 'special' form of authentication, you suffix your password with an underscore ( _ ). This can be changed in the [source code](legacy_2auth.lua).

### Original credentials
- Username: `john`
- Password: `secret`
- TOTP: `123456`

### New credentials
- Username: `john_`
- Password: `secret123456`

## Requirements
- [OpenResty](https://openresty.org/en/) or at least [lua-nginx-module](https://github.com/openresty/lua-nginx-module)
- [Authelia](https://github.com/authelia/authelia)

## Installation
Configure your nginx/openresty instance to use the `.conf` files in this repo. Customize as necessary.

## Notes
- Make sure `Set-Cookie` headers can reach the client through `auth_request` or the client will always create a new session and lose access after the TOTP expires. Check `auth_request_set` in [auth.conf](auth.conf)
- Make sure Authelia is aware of the real client IP or you may lock out your server on bruteforce attempts. Check `set_real_ip_from` in [auth.conf](auth.conf)

## TODO
- Handle multiple headers with same name
