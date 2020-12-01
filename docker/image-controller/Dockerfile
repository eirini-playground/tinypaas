# syntax = docker/dockerfile:experimental

ARG baseimage=scratch

FROM golang:1.15.5 as builder
WORKDIR /build/
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod vendor -trimpath -installsuffix cgo -o image-controller ./cmd/image-controller

FROM ${baseimage}
COPY --from=builder /build/image-controller /usr/local/bin/image-controller
USER 1001
ENTRYPOINT [ "/usr/local/bin/image-controller" ]
