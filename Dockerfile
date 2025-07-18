FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN cd cmd/server && \
    GOOS=linux CGO_ENABLED=0 go build -o server main.go

FROM scratch
COPY --from=builder /app/cmd/server/server .
COPY --from=builder /app/.env .
CMD ["./server"]