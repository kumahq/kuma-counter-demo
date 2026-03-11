ARG ARCH=amd64
FROM golang:1.25@sha256:bd1e2df4e6259b2bd5b1de0e6b22ca414502cd6e7276a5dd5dd414b65063be58

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
