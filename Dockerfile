FROM golang:1.10.4

RUN mkdir -p /go/src/github.com/gitmonster/faas-rancher/
WORKDIR /go/src/github.com/gitmonster/faas-rancher

COPY vendor     vendor
COPY handlers   handlers
COPY types      types
COPY rancher    rancher
COPY server.go  .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \  
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
#  -ldflags="-X main.CommitSHA=`git rev-parse HEAD`" \
  -o /tmp/faas-rancher .

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=0 /tmp/faas-rancher .
CMD ["./faas-rancher"]

