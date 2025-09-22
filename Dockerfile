ARG ARCH=amd64
FROM golang:1.25@sha256:8305f5fa8ea63c7b5bc85bd223ccc62941f852318ebfbd22f53bbd0b358c07e1

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
