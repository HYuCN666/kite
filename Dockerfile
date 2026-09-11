# ---- 构建后端 ----
FROM golang:1.27-alpine AS backend
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /kite ./cmd/kite

# ---- 构建前端 ----
FROM node:20-alpine AS frontend
WORKDIR /web
COPY web/package*.json ./
RUN npm install
COPY web/ .
RUN npm run build

# ---- 运行时 ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /kite ./kite
COPY --from=frontend /web/dist ./web/dist
EXPOSE 8080
ENTRYPOINT ["./kite"]
