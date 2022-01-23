FROM golang:1.17 AS builder

WORKDIR /src
COPY ./PingUtility .
RUN go mod download
RUN mkdir /old-logs
RUN CGO_ENABLED=1 GOOS=linux go build -o /app -a -ldflags '-linkmode external -extldflags "-static"' .

FROM scratch
COPY --from=builder /old-logs /old-logs
COPY --from=builder /app /app
ENTRYPOINT ["/app"]