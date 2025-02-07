FROM golang:1.23-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

FROM alpine:3.19 AS prod

WORKDIR /app

# Create necessary directories and set up non-root user
RUN adduser -D appuser && \
    mkdir -p statements templates && \
    chown -R appuser:appuser /app

# Copy files in a specific order
COPY --from=build /app/main .
COPY --from=build /app/templates ./templates/
COPY --from=build /app/statements ./statements/
COPY --from=build /app/README.md ./README.md

# Set correct permissions
RUN chown -R appuser:appuser /app && \
    chmod 755 statements && \
    chmod 644 README.md && \
    chmod 755 main

# Switch to non-root user
USER appuser

EXPOSE 8080

CMD ["./main"]