FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Tidy and vendor dependencies
RUN go mod tidy
RUN go mod vendor

COPY . .

# Build the application using vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o /arrivo-api ./cmd/api

# Stage 2: Create the final image
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /arrivo-api .

# Expose the port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/arrivo-api"]