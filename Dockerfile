FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/statements /app/statements
# Create statements directory and ensure proper permissions
RUN mkdir -p /app/statements && chmod 755 /app/statements
EXPOSE ${PORT}
CMD ["./main"]