FROM golang:1.11.5-alpine3.8 AS build
WORKDIR /build
COPY . ./

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") \  
  && CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo \
#  -ldflags="-X main.CommitSHA=`git rev-parse HEAD`" \
  -o /tmp/faas-rancher .

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
WORKDIR /root/

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

VOLUME ["/metastore"]

COPY --from=0 /tmp/faas-rancher .
CMD ["./faas-rancher"]

