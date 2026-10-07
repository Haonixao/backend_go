FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o main ./cmd/backend_go

FROM alpine:3.24
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 10001 appgroup && \
    adduser -D -u 10001 -G appgroup appuser
ENV TZ=Europe/Moscow
WORKDIR /app
COPY --from=builder /app/main .
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
CMD ["./main"]
