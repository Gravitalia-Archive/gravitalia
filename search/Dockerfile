FROM rust:alpine3.18 AS builder

RUN USER=root cargo new --bin search
WORKDIR /search

ENV     RUSTFLAGS="-C target-feature=-crt-static"
RUN     apk add -q --update-cache --no-cache build-base openssl-dev musl

COPY    . .
RUN    cargo build --release

FROM alpine:3.18 AS runtime

RUN apk update \
 && apk add --no-cache libssl1.1 musl-dev libgcc tini curl

COPY --from=builder /search/target/release/search /bin/search

EXPOSE 8890/tcp
ENTRYPOINT ["tini", "--"]
CMD     /bin/search
