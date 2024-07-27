FROM alpine:3.20

ARG TARGETOS
ARG TARGETARCH

LABEL maintainer="Bo-Yi Wu <appleboy.tw@gmail.com>" \
  org.label-schema.name="Sync Gitea Secrets" \
  org.label-schema.vendor="Bo-Yi Wu" \
  org.label-schema.schema-version="1.0"

LABEL org.opencontainers.image.source=https://github.com/appleboy/gitea-secret-sync
LABEL org.opencontainers.image.description="Sync Gitea Secrets"
LABEL org.opencontainers.image.licenses=MIT

RUN apk add --no-cache ca-certificates && \
  rm -rf /var/cache/apk/*

COPY release/${TARGETOS}/${TARGETARCH}/gitea-secret-sync /bin/

ENTRYPOINT ["/bin/gitea-secret-sync"]
