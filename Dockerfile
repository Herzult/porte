FROM golang:1.13-alpine AS builder
RUN apk add --no-cache git mercurial ca-certificates && update-ca-certificates
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY ./internal /app/internal 
COPY ./cmd /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o porte ./cmd/porte

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/porte .
EXPOSE 8080
ENTRYPOINT [ "/app/porte" ]