FROM golang:1.21.0 as builder

WORKDIR /app/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY main.go main.go
COPY prometheus/prometheus.go prometheus/prometheus.go
RUN CGO_ENABLED=0 go build -o /main


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /main /main
ARG DATE
LABEL org.opencontainers.image.version=${DATE}
ENTRYPOINT ["/main"]
