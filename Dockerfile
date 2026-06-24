ARG ARCH=amd64
FROM golang:1.26@sha256:8c5d338aa0da7e8b034efab738460793dc48d2aac8d497c8f8617d4ef964e0a7

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
