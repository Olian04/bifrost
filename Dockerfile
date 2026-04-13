FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Injected by CI (e.g. release workflow); local builds default to "unknown" (see cmd/bifrost/version).
ARG VERSION=unknown
ARG REVISION=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/lolocompany/bifrost/cmd/bifrost/version.Version=${VERSION} -X github.com/lolocompany/bifrost/cmd/bifrost/version.Revision=${REVISION} -X github.com/lolocompany/bifrost/cmd/bifrost/version.BuildTime=${BUILD_TIME}" -o /out/bifrost ./cmd/bifrost

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

RUN addgroup -S bifrost && adduser -S -G bifrost bifrost
USER bifrost

WORKDIR /app
COPY --from=build /out/bifrost /usr/local/bin/bifrost

ENTRYPOINT ["/usr/local/bin/bifrost"]
