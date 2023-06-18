FROM golang:1.20-alpine3.18 AS build

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN go build -o recommendation

FROM alpine:3.18 AS runtime

COPY --from=build /app/recommendation /app/recommendation

EXPOSE 8889
CMD [ "/app/recommendation" ]
