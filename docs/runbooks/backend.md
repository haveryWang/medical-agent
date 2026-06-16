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
- 火山引擎方舟 DeepSeek 对话 API
- 火山引擎方舟豆包向量模型 API

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

DEEPSEEK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DEEPSEEK_CHAT_MODEL=deepseek-v4-flash-260425
DEEPSEEK_API_KEY=请在这里填写你的火山引擎方舟密钥

QWEN_EMBEDDING_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
QWEN_EMBEDDING_API_KEY=请在这里填写你的火山引擎方舟密钥
QWEN_EMBEDDING_MODEL=doubao-embedding-vision-251215
QWEN_EMBEDDING_DIMENSION=2048
```

也可以只配置 `VOLCENGINE_API_KEY`，后端会把它作为 DeepSeek 对话和豆包向量模型的默认密钥。启动后也可以在前端顶栏点击“系统设置”打开模型配置弹窗，将火山引擎方舟 Base URL、模型名、API Key 和向量维度保存到 MongoDB。`doubao-embedding-vision-251215` 固定使用 2048 维。后端执行文档入库、向量生成和流式问答时会优先读取数据库配置；数据库尚未保存时使用上述环境变量作为兜底。

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
