# ---- Build Stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /build

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ipgeo .

# ---- Final Stage ----
FROM alpine:latest
WORKDIR /app

COPY --from=builder /build/ipgeo .
COPY data/ ./data/

EXPOSE 8081

ENTRYPOINT ["./ipgeo"]
