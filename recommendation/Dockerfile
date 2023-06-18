FROM golang:1.20-alpine3.18 AS build

RUN useradd -u 8877 nonroot
USER nonroot

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN GOOS=linux GOARCH=amd64 go build -o recommendation

FROM alpine:3.18 AS runtime

USER nonroot

COPY --from=build /app/recommendation /app/recommendation

EXPOSE 8889
CMD [ "/app/recommendation" ]
