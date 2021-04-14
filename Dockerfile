FROM golang:1.16-alpine as builder
WORKDIR /go/src/github.com/TimeBye/registry-manager
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build -o registry-manager

FROM alpine:3
RUN apk --no-cache add \
        jq \
        tini \
        curl \
        bash \
        screen \
        ca-certificates; \
    apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community skopeo; \
    skopeo -v
RUN cp /etc/apk/repositories /etc/apk/repositories.bak; \
        sed -i 's dl-cdn.alpinelinux.org mirrors.aliyun.com g' /etc/apk/repositories
COPY --from=builder /go/src/github.com/TimeBye/registry-manager/registry-manager /usr/bin/registry-manager
ADD https://raw.githubusercontent.com/containers/skopeo/master/default-policy.json /etc/containers/policy.json
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["registry-manager"]