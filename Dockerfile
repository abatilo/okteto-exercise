FROM golang:1.18.4-alpine as build

# SSL Certificates
RUN apk --no-cache add ca-certificates

WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download -x

COPY cmd ./cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/podlist cmd/podlist/main.go

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/podlist /usr/local/bin/podlist

EXPOSE 8080
ENTRYPOINT ["podlist"]
