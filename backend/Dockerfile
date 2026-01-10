FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git build-base ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# build both api and worker binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/worker ./cmd/worker

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin /app/bin
EXPOSE 3000
USER 1000:1000
ENTRYPOINT ["/app/bin/api"]
