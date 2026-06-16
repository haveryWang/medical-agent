# 后端启动和构建说明

后端目录：

```text
backend/
```

技术栈：

- Go
- chi router / chi middleware
- MongoDB 官方驱动
- Qdrant HTTP API
- DeepSeek Chat API
- Qwen3-Embedding API

## 环境变量

复制模板：

```bash
cd backend
cp .env.example .env
```

关键配置：

```text
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=medical_agent
QDRANT_URL=http://localhost:6333
QDRANT_COLLECTION=medical_agent_chunks

DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_CHAT_MODEL=deepseek-v4-flash
DEEPSEEK_API_KEY=请在这里填写你的DeepSeek密钥

QWEN_EMBEDDING_BASE_URL=请在这里填写Qwen3-Embedding服务地址
QWEN_EMBEDDING_API_KEY=请在这里填写Qwen3-Embedding密钥
QWEN_EMBEDDING_MODEL=Qwen3-Embedding
QWEN_EMBEDDING_DIMENSION=1024
```

## 启动依赖服务

```bash
docker compose up -d mongodb qdrant
```

## 初始化 Qdrant

```bash
cd backend
sh scripts/qdrant-init.sh
```

## 启动后端

```bash
cd backend
go mod download
set -a
. ./.env
set +a
go run ./cmd/server
```

访问健康检查：

```bash
curl http://localhost:8080/health
```

## 构建

```bash
cd backend
go build ./cmd/server
```

## 数据库说明

MongoDB 集合和索引脚本：

```text
backend/scripts/mongo-init.js
```

后端启动时也会自动调用 `EnsureIndexes` 创建索引，并在空库时插入演示账号和知识库数据。

Qdrant collection 初始化脚本：

```text
backend/scripts/qdrant-init.sh
```

## 后端代码结构

```text
backend/
  cmd/server/main.go              # 服务入口
  internal/app/app.go             # 依赖初始化和后台 ingestion worker
  internal/httpapi/router.go      # chi 路由、CORS、中间件挂载
  internal/httpapi/middleware.go  # 请求 ID、登录态、权限校验
  internal/httpapi/*_handlers.go  # auth、knowledge、ingestion、chat、health handlers
  internal/httpapi/requests.go    # 请求 DTO
  internal/httpapi/response.go    # JSON/SSE 响应工具
  internal/store/mongo.go         # Mongo 连接、索引、公共类型
  internal/store/auth.go          # 用户和会话数据访问
  internal/store/knowledge.go     # 知识库和文档查询
  internal/store/ingestion.go     # 文档入库、任务、chunk 写入
  internal/store/chat.go          # 会话和消息数据访问
  internal/store/rag.go           # RAG chunk 查询
  internal/store/audit.go         # 审计日志
  internal/store/seed.go          # 空库初始化数据
```

`POST /api/v1/knowledge-bases`、`PATCH /api/v1/knowledge-bases/{id}`、文档上传和任务重试现在通过后端权限中间件校验 `knowledge:write`。消息详情接口会校验消息所属会话是否属于当前用户。
