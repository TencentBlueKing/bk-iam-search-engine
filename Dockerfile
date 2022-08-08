FROM golang:1.18 AS builder

COPY ./ /app
WORKDIR /app

ARG BINARY=iam-search-engine

RUN make build && chmod +x ${BINARY}
RUN mkdir -p /tmp/app/logs
RUN cp ${BINARY} /tmp/app

FROM debian:bullseye-slim
COPY --from=builder /tmp/app /app

CMD ["/app/iam-search-engine", "-c", "/app/config.yaml"]
