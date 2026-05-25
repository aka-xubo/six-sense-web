遵循[claude](CLAUDE.md)

## 本地服务排障经验

- 后端已替换为 Go 实现，默认启动命令：
  ```bash
  cd backend-go
  GOCACHE=/private/tmp/go-build go run ./cmd/server
  ```
- 重启 8000 前优先检查端口：
  ```bash
  lsof -iTCP:8000 -sTCP:LISTEN -n -P
  ```
- 普通 `kill <pid>` 无效时，可强制释放开发端口：
  ```bash
  lsof -tiTCP:8000 -sTCP:LISTEN | xargs kill -9
  ```
