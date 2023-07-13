FROM rust:alpine3.18 AS builder

RUN USER=root cargo new --bin recommendation
WORKDIR /recommendation

ENV     RUSTFLAGS="-C target-feature=-crt-static"
RUN     apk add -q --update-cache --no-cache build-base openssl-dev musl

COPY ./Cargo.toml ./Cargo.toml
RUN cargo build --release \
 && rm src/*.rs

COPY ./src ./src
RUN rm ./target/release/deps/recommendation* \
 && cargo build --release

FROM alpine:3.18 AS runtime

RUN apk update \
 && apk add --no-cache libssl1.1 musl-dev libgcc tini curl

COPY --from=builder /recommendation/target/release/recommendation /bin/recommendation

EXPOSE 8889/tcp
ENTRYPOINT ["tini", "--"]
CMD     /bin/recommendation
