FROM rust:1.70 as build

RUN USER=root cargo new --bin search
WORKDIR /search

COPY ./Cargo.toml ./Cargo.toml

RUN apt-get update && apt-get install -y --no-install-recommends libssl-dev pkg-config

RUN cargo build --release \
 && rm src/*.rs

COPY ./src ./src

RUN rm ./target/release/deps/search* \
 && cargo build --release

FROM rust:1.70-slim-buster

COPY --from=build /search/target/release/search .

EXPOSE 8890
CMD ["./search"]
