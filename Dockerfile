FROM --platform=linux/amd64 golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ssh-portfolio .

# Minimal runtime — no OS, just the static binary.
FROM scratch

WORKDIR /app
COPY --from=builder /app/ssh-portfolio /app/ssh-portfolio

EXPOSE 22

CMD ["/app/ssh-portfolio"]
