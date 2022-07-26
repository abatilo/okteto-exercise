FROM golang:1.18.4-alpine as build
ENV GOFLAGS=-mod=vendor

# SSL Certificates
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
COPY internal ./internal/
COPY vendor ./vendor/
COPY cmd ./cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/podlist cmd/podlist/main.go

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/podlist /usr/local/bin/podlist

EXPOSE 8080
EXPOSE 8081
ENTRYPOINT ["podlist"]
