ARG ARCH=amd64
FROM golang:1.26@sha256:079e59808d2d252516e27e3f3a9c003740dee7f75e55aa71528766d52bcfc16a

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
