FROM golang:1.22-alpine AS builder
RUN apk add --update build-base git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/app .
EXPOSE 8080
CMD ["/app/app"]
