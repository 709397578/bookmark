# 多阶段构建 - 后端
# Build arg: CGO_ENABLED=0 for PostgreSQL, CGO_ENABLED=1 for SQLite
ARG CGO_ENABLED=0
# Build arg: INCLUDE_CHROME=0 to skip, INCLUDE_CHROME=1 to install Chromium (for snapshots)
ARG INCLUDE_CHROME=0

FROM golang:1.25-alpine AS backend-builder

WORKDIR /app/backend

# 安装必要的依赖
RUN apk add --no-cache git
# SQLite 需要 CGO 和 gcc
ARG CGO_ENABLED
RUN if [ "$CGO_ENABLED" = "1" ]; then apk add --no-cache gcc musl-dev; fi
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
# 复制go.mod和go.sum
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制源代码
COPY backend/ .

# 构建应用
RUN CGO_ENABLED=$CGO_ENABLED GOOS=linux go build -o /pintree-backend cmd/main.go

# 最终镜像
FROM alpine:latest

WORKDIR /root/

# 安装ca-certificates（用于HTTPS请求）
RUN apk --no-cache add ca-certificates

# SQLite 模式需要保留 CGO 运行时依赖
ARG CGO_ENABLED
RUN if [ "$CGO_ENABLED" = "1" ]; then apk add --no-cache libstdc++; fi

# Chromium 快照支持（可选，增加约 200MB）
ARG INCLUDE_CHROME
RUN if [ "$INCLUDE_CHROME" = "1" ]; then \
        apk add --no-cache \
            chromium \
            ttf-freefont \
            font-noto \
            font-noto-cjk \
            font-noto-extra; \
    fi

# 从builder阶段复制二进制文件
COPY --from=backend-builder /pintree-backend .

# 复制.env文件（如果存在）
COPY backend/.env .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./pintree-backend"]
