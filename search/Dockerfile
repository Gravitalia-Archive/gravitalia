FROM rust:1.69 as build

RUN USER=root cargo new --bin search
WORKDIR /search

COPY ./Cargo.toml ./Cargo.toml

RUN cargo build --release \
 && rm src/*.rs

RUN apt-get update && apt-get install -y --no-install-recommends libssl-dev pkg-config protobuf-compiler

COPY ./src ./src

RUN rm ./target/release/deps/search* \
 && cargo build --release

FROM alpine:latest

COPY --from=build /search/target/release/search .

EXPOSE 8890
CMD ["./search"]
