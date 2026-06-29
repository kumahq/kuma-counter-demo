ARG ARCH=amd64
FROM golang:1.26@sha256:f96cc555eb8db430159a3aa6797cd5bae561945b7b0fe7d0e284c63a3b291609

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
