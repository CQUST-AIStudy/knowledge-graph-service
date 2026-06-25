# 多阶段构建（参考 LeetCodeClaw）

ARG GO_IMAGE=docker.m.daocloud.io/golang:1.25-alpine
ARG RUNTIME_IMAGE=docker.m.daocloud.io/alpine:3.22

FROM ${GO_IMAGE} AS builder

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}
ENV CGO_ENABLED=0

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/knowledge-graph-api ./cmd/knowledge-graph-api

FROM ${RUNTIME_IMAGE}

RUN apk add --no-cache ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    adduser -D -u 1000 app

COPY --from=builder /out/knowledge-graph-api /app/knowledge-graph-api

ENV KG_ADDR=:10171
ENV TZ=Asia/Shanghai

EXPOSE 10171
USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -q -O- http://127.0.0.1:10171/ready || exit 1

ENTRYPOINT ["/app/knowledge-graph-api"]
