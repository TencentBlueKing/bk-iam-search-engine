FROM golang:1.22-bullseye AS builder

COPY ./ /app
WORKDIR /app

ARG BINARY=iam-search-engine

RUN make build && chmod +x ${BINARY}
RUN mkdir -p /tmp/app/logs
RUN cp ${BINARY} /tmp/app

FROM tencentos/tencentos3-minimal
COPY --from=builder /tmp/app /app

CMD ["/app/iam-search-engine", "-c", "/app/config.yaml"]
