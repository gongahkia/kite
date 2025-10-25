# syntax=docker/dockerfile:1

FROM nimlang/nim:2.0.4 as build
WORKDIR /src
COPY kite.nimble .
COPY src ./src
RUN nimble install -y
RUN nimble build -y -d:release

FROM debian:stable-slim
WORKDIR /app
COPY --from:build /root/.nimble/bin/kite /app/kite
ENV METRICS_PORT=8000 METRICS_PATH=/metrics ENABLE_METRICS=true LOG_LEVEL=INFO
EXPOSE 8000
ENTRYPOINT ["/app/kite"]

