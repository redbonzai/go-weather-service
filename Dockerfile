# ---- Build stage ----
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN apk add --no-cache tzdata

# Cache Go dependencies without requiring go.sum to exist
COPY go.mod ./
RUN go mod download

# Now copy the rest of the source
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /bin/weather-service ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/weather-service /usr/local/bin/weather-service
RUN adduser -D -H -s /sbin/nologin appuser
USER appuser
EXPOSE 8080
ENV COLD_LT_F=50 HOT_GE_F=80
ENTRYPOINT ["weather-service"]
