FROM golang:1.16 AS builder

# install oapi-codegen
COPY scripts/gen/install.sh /install-tools.sh
RUN bash /install-tools.sh

# add bot api spec
COPY scripts/gen/telegram-bot-api.json /app/

RUN PLATFORMS="telegram" && \
    for platform in ${PLATFORMS}; do \
      mkdir -p "/app/${platform}" && \
      oapi-codegen \
        -package "${platform}" \
        -o "/app/${platform}/openapi.go" \
        -generate types,client \
        "/app/${platform}-bot-api.json" ; \
    done

FROM scratch

COPY --from=builder /app/telegram /telegram
