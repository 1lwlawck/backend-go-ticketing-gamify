# Build stage
FROM golang:1.23-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.20
WORKDIR /app
ENV PORT=8080
EXPOSE 8080
COPY --from=build /app/server /app/server
COPY --from=build /app/.env.example /app/.env.example
CMD ["/app/server"]
