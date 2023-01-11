FROM golang:1.19-buster AS builder

WORKDIR $GOPATH/src/kafka-consumer/
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/kafka-consumer

# Smallish image for actual running
FROM gcr.io/distroless/base-debian10
COPY --from=builder /go/bin/kafka-consumer /go/bin/kafka-consumer


EXPOSE 8080
ENTRYPOINT ["/go/bin/kafka-consumer"]