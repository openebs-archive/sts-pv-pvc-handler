FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

WORKDIR /app

COPY *.go ./
COPY pkg ./pkg
COPY Makefile ./
COPY tests ./tests

RUN go build -o /lister-sa

EXPOSE 8080

CMD [ "/lister-sa" ]
