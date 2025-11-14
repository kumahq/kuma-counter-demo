ARG ARCH=amd64
FROM golang:1.25@sha256:e68f6a00e88586577fafa4d9cefad1349c2be70d21244321321c407474ff9bf2

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
