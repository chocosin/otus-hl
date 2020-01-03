FROM golang:1.13 as builder

WORKDIR /build

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build .

FROM alpine
COPY templates /templates
COPY migrations /migrations
COPY --from=builder /build/social .
CMD ["./social"]
