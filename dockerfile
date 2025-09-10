# Build stage
FROM golang:1.22 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o /app/server ./cmd/service

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/server /app/server
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/server"]