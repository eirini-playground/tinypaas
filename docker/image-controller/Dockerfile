# syntax = docker/dockerfile:experimental

ARG baseimage=scratch

FROM golang:1.15.5 as builder
WORKDIR /build/
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod vendor -trimpath -installsuffix cgo -o image-controller ./cmd/image-controller
ARG GIT_SHA
RUN if [ -z "$GIT_SHA" ]; then echo "GIT_SHA not set"; exit 1; else : ; fi

FROM ${baseimage}
COPY --from=builder /build/image-controller /usr/local/bin/image-controller
USER 1001
ENTRYPOINT [ "/usr/local/bin/image-controller" ]
ARG GIT_SHA
LABEL org.opencontainers.image.revision=$GIT_SHA \
      org.opencontainers.image.source=https://github.com/jimmykarily/tinypaas.git
