FROM golang:1.24-alpine3.21 AS builder

WORKDIR /prouter

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./

ARG VERSION

RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=${VERSION}"

RUN mkdir -p /public

FROM gcr.io/distroless/static-debian12 AS runner

COPY --from=builder /prouter/prouter /prouter

COPY --from=builder /public /public

ENTRYPOINT ["/prouter", "--serve", "/public"]
