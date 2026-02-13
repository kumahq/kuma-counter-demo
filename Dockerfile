ARG ARCH=amd64
FROM golang:1.25@sha256:85c0ab0b73087fda36bf8692efe2cf67c54a06d7ca3b49c489bbff98c9954d64

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
