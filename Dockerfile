
FROM golang:1.14 as builder

WORKDIR /app

COPY . .

RUN go build .

FROM golang:1.14

LABEL maintainer="dustin.k.lo@nasa.jpl.gov"

COPY --from=builder /app/redis-pubsub-websocket /redis-pubsub-websocket

EXPOSE 8000

ENTRYPOINT ["/redis-pubsub-websocket"]
CMD ["-h"]
