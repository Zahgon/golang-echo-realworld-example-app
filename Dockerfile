#
# 1. Build Container
#
FROM golang:1.25 AS build

# No GOOS/GOARCH here on purpose. In a docker build the build stage already
# runs on the target platform, so pinning an architecture turns every build on
# a different host into a cross compile - and go then disables cgo by default,
# which silently swaps the sqlite3 driver for a stub that fails at runtime
# rather than at build time. CGO_ENABLED is set explicitly so that degrading
# can never happen quietly; use docker's --platform to choose an architecture.
ENV GO111MODULE=on \
    CGO_ENABLED=1

RUN mkdir -p /src

# First add modules list to better utilize caching
COPY go.sum go.mod /src/

WORKDIR /src

# Download dependencies
RUN go mod download

COPY . /src

# Build components.
# Put built binaries and runtime resources in /app dir ready to be copied over or used.
# go build -o writes exactly where it is told. go install placed the binary in
# an architecture suffixed directory whenever the build cross compiled, which
# broke the copy that followed on any host that was not linux/amd64.
RUN mkdir -p /app && \
    go build -installsuffix cgo -ldflags="-w -s" -o /app/golang-gin-realworld-example-app .

#
# 2. Runtime Container
#
FROM debian:stable-slim

LABEL maintainer="Sina Saeidi <xesina@gmail.com>"

ENV TZ=Asia/Tehran \
    PATH="/app:${PATH}"

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    sqlite3 \
    tzdata \
    ca-certificates \
    bash \
    && \
    rm -rf /var/lib/apt/lists/* && \
    cp --remove-destination /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

WORKDIR /app

COPY --from=build /app /app/

EXPOSE 8585

CMD ["./golang-gin-realworld-example-app"]
