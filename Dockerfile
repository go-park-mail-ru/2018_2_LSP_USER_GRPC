FROM golang:alpine

RUN apk add --no-cache git

ADD . /go/src/github.com/go-park-mail-ru/2018_2_LSP_AUTH_GRPC

RUN cd /go/src/github.com/go-park-mail-ru/2018_2_LSP_AUTH_GRPC && go get ./...

RUN go install github.com/go-park-mail-ru/2018_2_LSP_AUTH_GRPC

ENTRYPOINT /go/bin/2018_2_LSP_AUTH_GRPC

EXPOSE 8080