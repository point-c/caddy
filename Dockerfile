FROM caddy:2.7.6-builder AS builder

RUN xcaddy build \
    --with github.com/point-c/caddy/module@v0.1.2
FROM caddy:2.7.6

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
