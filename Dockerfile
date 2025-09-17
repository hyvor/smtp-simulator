FROM golang:1.22-alpine AS builder
WORKDIR /app
# Install git (needed for go mod download)
RUN apk add --no-cache git 
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smtp-simulator .

# --- 
FROM alpine:latest

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/smtp-simulator .

RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 25

# Run the SMTP simulator
CMD ["./smtp-simulator"]
