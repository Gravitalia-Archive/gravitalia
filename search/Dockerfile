FROM rust:1.69 as build

RUN USER=root cargo new --bin search
WORKDIR /search

COPY ./Cargo.toml ./Cargo.toml

RUN cargo build --release \
 && rm src/*.rs

COPY ./src ./src

RUN rm ./target/release/deps/search* \
 && cargo build --release

FROM alpine:3.18

COPY --from=build /search/target/release/search .

EXPOSE 8890
CMD ["./search"]
