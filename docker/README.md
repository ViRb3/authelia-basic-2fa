# Docker setup

This is a complete Docker setup to demonstrate how to use this project with Authelia and nginx. Please check [docker-compose.yml](docker-compose.yml) and follow it to understand how everything works.

- The nginx Authelia configuration was taken from [the official documentation](https://github.com/authelia/authelia/blob/b20f62b0151c2ec0c35003746ff69f4be979959d/docs/deployment/supported-proxies/nginx.md).

- Make sure to change `auth.website.com` to your domain in the nginx conf files.
