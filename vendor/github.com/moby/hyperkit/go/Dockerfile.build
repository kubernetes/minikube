FROM golang:1.8-alpine
RUN apk update && apk add --no-cache make git
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# The project sources
VOLUME ["/go/src/github.com/moby/hyperkit/go"]
WORKDIR /go/src/github.com/moby/hyperkit/go

ENTRYPOINT ["make"]
