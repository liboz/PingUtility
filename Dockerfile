FROM golang:1.22 AS builder

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY . .
# need CGO for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o /app -a -ldflags '-w -s -linkmode external -extldflags -static' ./Database

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app /pingutility-database
ENTRYPOINT ["/pingutility-database"]