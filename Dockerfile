# syntax=docker/dockerfile:1.7
# Multi-stage build → tiny final image (~15 MB).

FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/animetsu-api ./cmd/server

# Final: distroless static. PORT is overridable; default 8080.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/animetsu-api /animetsu-api
ENV PORT=8080 \
    RATE_LIMIT_RPM=120 \
    LOG_LEVEL=info
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/animetsu-api"]
