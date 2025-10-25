# syntax=docker/dockerfile:1

FROM nimlang/nim:2.0.4 as build
WORKDIR /src
COPY kite.nimble .
COPY src ./src
RUN nimble install -y && \
    nimble build -y -d:release --opt:speed

FROM debian:stable-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /root/.nimble/bin/kite /app/kite
ENV METRICS_PORT=8000 \
    METRICS_PATH=/metrics \
    ENABLE_METRICS=true \
    LOG_LEVEL=INFO
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/kite", "--version"]
ENTRYPOINT ["/app/kite"]
