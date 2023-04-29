FROM golang:1.20 AS build
EXPOSE 8888

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o gravitalia
CMD [ "/app/gravitalia" ]
