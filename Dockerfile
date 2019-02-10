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

ENV HOST=0.0.0.0
ENV PORT=8100
ENV CONCURRENCY=1
ENV SLEEP=1
ENV TARGETS=

ENTRYPOINT goload
