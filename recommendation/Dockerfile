FROM golang:1.20 AS build
EXPOSE 8889

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN GOOS=linux GOARCH=amd64 go build -o recommendation
CMD [ "/app/recommendation" ]
