local cjson = require "cjson"

local sess = ngx.var.cookie_authelia_session
local auth_header = ngx.req.get_headers()["authorization"]

--- disable basic auth so it doesn't take precedence over 2FA
ngx.req.clear_header("authorization")

-- first try to see if the session is still valid
if sess ~= nil then
    local res = ngx.location.capture("/api/verify", {method = ngx.HTTP_GET})
    if res.status == ngx.HTTP_OK then
        ngx.log(ngx.ERR, "valid session")
        ngx.exit(ngx.HTTP_OK)
    end
end

--- re-enable basic auth so it's not lost in case the script returns
ngx.req.set_header("Authorization", auth_header)

-- if not, try to use customized auth basic
if auth_header == nil then
    ngx.log(ngx.ERR, "no auth header")
    return
end

local auth_split = {}
for i in string.gmatch(auth_header, "%S+") do
    auth_split[#auth_split + 1] = i
end
if #auth_split ~= 2 then
    ngx.log(ngx.ERR, "unrecognized auth header format")
    return
end

if string.lower(auth_split[1]) ~= "basic" then
    ngx.log(ngx.ERR, "not auth basic")
    return
end

local auth_decoded = ngx.decode_base64(auth_split[2])
if auth_decoded == nil then
    ngx.log(ngx.ERR, "not base64 auth basic")
    return
end

local user_pass = {}
for i in string.gmatch(auth_decoded, "[^:]+") do
    user_pass[#user_pass + 1] = i
end
if #user_pass ~= 2 then
    ngx.log(ngx.ERR, "unrecognized auth basic format")
    return
end

if string.len(user_pass[1]) < 2 or string.sub(user_pass[1], -1) ~= "_" then
    ngx.log(ngx.ERR, "unrecognized user format or not special auth")
    return
end
if string.len(user_pass[2]) < 7 then
    ngx.log(ngx.ERR, "unrecognized pass format")
    return
end

local username = string.sub(user_pass[1], 0, #user_pass[1] - 1)
local password = string.sub(user_pass[2], 0, #user_pass[2] - 6)
local totp = string.sub(user_pass[2], -6)

--- disable basic auth so it doesn't take precedence over 2FA
ngx.req.clear_header("authorization")

local res =
    ngx.location.capture(
    "/api/firstfactor",
    {
        method = ngx.HTTP_POST,
        body = '{"username":"' .. username .. '","password":"' .. password .. '","keepMeLoggedIn":false}'
    }
)

-- check request status
if res.status ~= ngx.HTTP_OK or res.truncated then
    ngx.log(ngx.ERR, "failed first factor")
    ngx.exit(ngx.ERROR)
end

-- check response message status
local msg = cjson.decode(res.body)
if msg["status"] ~= "OK" then
    ngx.log(ngx.ERR, "failed first factor (2)")
    ngx.exit(ngx.ERROR)
end

-- use the session cookie returned by first factor auth
-- the cookie won't change, so no need to do again after second factor
for k, v in pairs(res.header) do
    if string.lower(k) == "set-cookie" then
        ngx.log(ngx.ERR, "got session cookie")
        -- set server-side for second factor
        ngx.req.set_header("Cookie", v)
        -- return to client
        ngx.header[k] = v
    end
end

local res =
    ngx.location.capture(
    "/api/secondfactor/totp",
    {
        method = ngx.HTTP_POST,
        body = '{"token":"' .. totp .. '"}'
    }
)

-- check request status
if res.status ~= ngx.HTTP_OK or res.truncated then
    ngx.log(ngx.ERR, "failed second factor")
    ngx.exit(ngx.ERROR)
end

-- check response message status
local msg = cjson.decode(res.body)
if msg["status"] ~= "OK" then
    ngx.log(ngx.ERR, "failed second factor (2)")
    ngx.exit(ngx.ERROR)
end

-- check if we should grant access
local res = ngx.location.capture("/api/verify", {method = ngx.HTTP_GET})
if res.status ~= ngx.HTTP_OK then
    ngx.log(ngx.ERR, "failed auth verify")
    ngx.exit(ngx.ERROR)
end

-- bypasses proxy_pass
ngx.exit(ngx.HTTP_OK)
