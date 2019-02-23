# Building

FROM golang:1.11 as builder

WORKDIR $GOPATH/src/goload
COPY vendor/ $GOPATH/src/
COPY *.go ./
RUN go test
RUN CGO_ENABLED=0 GOOS=linux go build

# Running

FROM alpine

COPY --from=builder /go/src/goload/goload /usr/bin/goload
COPY entrypoint.sh /usr/sbin
RUN chmod +x /usr/sbin/entrypoint.sh

ENV HOST 0.0.0.0
ENV PORT 9115
ENV LOG_LEVEL ERROR
ENV CONCURRENCY 1
ENV SLEEP 1
ENV REPEAT -1
ENV TARGETS ""

ENTRYPOINT ["entrypoint.sh"]
