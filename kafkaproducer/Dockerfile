FROM golang:1.19-buster AS builder

WORKDIR $GOPATH/src/kafka-producer/
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/kafka-producer

# Smallish image for actual running
FROM gcr.io/distroless/base-debian10
COPY --from=builder /go/bin/kafka-producer /go/bin/kafka-producer


EXPOSE 8080
ENTRYPOINT ["/go/bin/kafka-producer"]