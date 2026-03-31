FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /poker-equity ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /poker-equity /poker-equity

EXPOSE 8080

ENV PORT=8080

CMD ["/poker-equity"]
