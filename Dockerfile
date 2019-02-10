FROM golang:1.11

ENV HOST=localhost
ENV PORT=8100
ENV CONCURRENCY=1
ENV SLEEP=1
ENV REQUESTS=

WORKDIR $GOPATH/src/goload
COPY vendor/ $GOPATH/src/
COPY *.go ./
RUN go test && go build && go install

ENTRYPOINT goload
