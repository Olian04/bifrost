FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Injected by CI (e.g. release workflow); local builds default to "unknown" (see cmd/bifrost/version.go).
ARG VERSION=unknown
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" -o /out/bifrost ./cmd/bifrost

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

RUN addgroup -S bifrost && adduser -S -G bifrost bifrost
USER bifrost

WORKDIR /app
COPY --from=build /out/bifrost /usr/local/bin/bifrost

ENTRYPOINT ["/usr/local/bin/bifrost"]
