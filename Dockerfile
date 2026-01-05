# Stage 1: Builder
FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN make build-linux

# Stage 2: Final Image
FROM alpine:latest

# Install ffmpeg, ca-certificates, timezone data
RUN apk add --no-cache ffmpeg tzdata ca-certificates && update-ca-certificates

COPY --from=builder /app/bin/ewallet ./ewallet
RUN chmod +x ./ewallet

EXPOSE 8081

CMD ["./ewallet"]
