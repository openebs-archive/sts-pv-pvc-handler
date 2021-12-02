FROM ksraj123/sts-stale-pvc-cleaner-base:0.1.0

WORKDIR /app

COPY *.go ./
COPY pkg ./pkg
COPY Makefile ./
COPY tests ./tests

RUN go build -o /lister-sa

EXPOSE 8080

CMD [ "/lister-sa" ]
