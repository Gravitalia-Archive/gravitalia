FROM golang:1.19 AS build
EXPOSE 8888

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /gravitalia

CMD [ "/gravitalia" ]
