FROM golang:1.20-alpine3.18 AS build

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN GOOS=linux GOARCH=amd64 go build -o rest

FROM alpine:3.18 AS runtime

COPY --from=build /app/recommendation /app/rest

EXPOSE 8888
CMD [ "/app/rest" ]
