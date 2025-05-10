FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/server/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -g '' appuser

WORKDIR /app
RUN mkdir -p /app/uploads

COPY --from=builder /app/app /app/
COPY --from=builder /app/config/config.yaml /app/config/
COPY --from=builder /app/docs /app/docs

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

ENV CONFIG_PATH="/app/config/config.yaml"

CMD ["/app/app"]