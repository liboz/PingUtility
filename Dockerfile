FROM golang:1.17 AS builder

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /app -a -ldflags '-w -s' ./Database

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /app
ENTRYPOINT ["/app"]