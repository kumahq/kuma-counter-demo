ARG ARCH=amd64
FROM golang:1.25@sha256:6ea52a02734dd15e943286b048278da1e04eca196a564578d718c7720433dbbe

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
