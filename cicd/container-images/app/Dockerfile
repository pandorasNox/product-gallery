FROM docker.io/golang:1.21.3-alpine3.18 AS builder

WORKDIR /workdir

COPY . .

RUN go build -o=./importer cmd/import/main.go
RUN go build -o=./server-api main.go

# =============================================================================

FROM docker.io/alpine:3.18.4 AS importer

COPY --from=builder /workdir/importer /importer

# =============================================================================

FROM docker.io/library/node:20.10-alpine3.18 as css-generator

ENV WORKDIR /workdir
WORKDIR ${WORKDIR}

# RUN npm install -D tailwindcss
# RUN npm install --save-dev tailwindcss
RUN npm install --global tailwindcss

COPY configs/tailwind.config.js ${WORKDIR}/tailwind.config.js
COPY templates /templates

RUN npx tailwindcss -i /templates/css/input.css -o /static/assets/css/generated/output.css

# =============================================================================

FROM docker.io/alpine:3.18.4 AS server

# packages:
#   * mailcap
#     desc: Helper application and MIME type associations for file types
#     info: https://pkgs.alpinelinux.org/package/edge/main/aarch64/mailcap
#     reason: see https://stackoverflow.com/questions/54835510/getting-mime-type-text-plain-error-in-golang-while-serving-css
RUN apk add --no-cache \
    mailcap

COPY --from=builder /workdir/server-api /server-api
COPY ./templates /templates
COPY ./static /static

COPY --from=css-generator /static/assets/css/generated/output.css /static/assets/css/generated/output.css

ENTRYPOINT [ "/server-api" ]
