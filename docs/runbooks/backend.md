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

复盘笔记和政策文件库使用独立集合：

```text
review_notes
review_note_exports
policy_documents
policy_import_batches
```

这些集合不会进入知识库文档入库、chunk 生成或 Qdrant 向量索引。

复盘笔记列表使用服务端分页；生成文档时只导出前端选中的记录，并在 `review_note_exports.content` 保存 Markdown 快照。后台最多保留最近 5 条导出文档，删除原始复盘记录不会影响已生成文档的再次下载。政策文件库支持删除单条 `policy_documents` 记录。

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
  internal/policy/                 # 政策固定分类和 Excel 导入解析
  internal/reviewnotes/            # 复盘笔记 Markdown 导出工具
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

## 复盘笔记和政策文件库

- 复盘笔记创建与 Markdown 导出需要 `review_notes:write` 权限。
- 政策库 Excel 导入需要 `policy:write` 权限。
- 默认空库管理员账号已包含上述权限。
- 复盘笔记导出接口会把当前未导出记录合并为 Markdown，并将这些记录标记为已导出。
- 政策库导入仅支持 `.xlsx`，导入模板表头为标题、摘要、解读、日期、分类标签；后端兼容标题、摘要、日期、分类字段，并可读取可选解读字段。
- 政策列表接口返回分类和月份聚合，可通过 `category` 与 `date` 参数筛选。
- 支持的政策分类固定为：国家医学中心、科技创新、医疗服务、医保医药、数智治理、改革监管、其他。
