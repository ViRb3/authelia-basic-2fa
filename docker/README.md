# Docker setup
This is a complete Docker setup to demonstrate how to use this project with Authelia and nginx. Please check [docker-compose.yml](docker-compose.yml) and follow it to understand how everything works.

- The nginx Authelia configuration was taken from [the official documentation](https://github.com/authelia/authelia/blob/master/docs/deployment/supported-proxies/nginx.md).

- Make sure to change `auth.website.com` to your domain in [_AUTH.conf](nginx/data/conf.d/_AUTH.conf).

- Make sure to change the `VERSION` variable in this project's [Dockerfile](authelia-basic-2fa/Dockerfile).
