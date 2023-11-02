FROM docker.io/golang:1.21.3-alpine3.18 AS builder

WORKDIR /workdir

COPY . .

RUN go build -o=./importer cmd/import/main.go
RUN go build -o=./server-api main.go

# =============================================================================

FROM docker.io/alpine:3.18.4 AS importer

COPY --from=builder /workdir/importer /importer

# =============================================================================

FROM docker.io/alpine:3.18.4 AS server

COPY --from=builder /workdir/server-api /server-api
COPY ./templates /templates

ENTRYPOINT [ "/server-api" ]
