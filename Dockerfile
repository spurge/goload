FROM golang:1.11

ENV HOST=localhost
ENV PORT=8100

WORKDIR $GOPATH/src/goload
COPY vendor/ $GOPATH/src/
COPY *.go ./
RUN go test && go build && go install

ENTRYPOINT goload
