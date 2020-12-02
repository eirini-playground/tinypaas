# syntax = docker/dockerfile:experimental

ARG baseimage=scratch

FROM golang:1.15.5 as builder
WORKDIR /tinypaas/
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod vendor -trimpath -installsuffix cgo -o eirini-routing ./routing/cmd
ARG GIT_SHA
RUN if [ -z "$GIT_SHA" ]; then echo "GIT_SHA not set"; exit 1; else : ; fi

FROM ${baseimage}
COPY --from=builder /tinypaas/eirini-routing /usr/local/bin/eirini-routing
USER 1001
ENTRYPOINT [ "/usr/local/bin/eirini-routing" ]
ARG GIT_SHA
LABEL org.opencontainers.image.revision=$GIT_SHA \
      org.opencontainers.image.source=https://github.com/jimmykarily/tinypaas.git
