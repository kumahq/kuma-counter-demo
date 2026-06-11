ARG ARCH=amd64
FROM golang:1.26@sha256:87a41d2539e5671777734e91f467499ed5eafb1fb1f77221dff2744db7a51775

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build ./... -o /kuma-counter-demo

FROM --platform=linux/${ARCH} distroless

CMD ["/kuma-counter-demo"]
