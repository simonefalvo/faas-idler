FROM golang:1.15-alpine AS builder
WORKDIR /go/src/github.com/smvfal/faas-idler
COPY . .
RUN go build -o faas-idler

FROM alpine
WORKDIR /root/
COPY --from=builder /go/src/github.com/smvfal/faas-idler/faas-idler .
CMD ["./faas-idler"]