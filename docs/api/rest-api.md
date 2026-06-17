# REST API 对接文档

Base URL:

```text
http://localhost:8080/api/v1
```

除登录接口外，所有接口都需要请求头：

```text
Authorization: Bearer <token>
```

创建/更新知识库、上传文档、重试入库任务需要当前用户拥有 `knowledge:write` 权限。复盘笔记创建/导出需要 `review_notes:write` 权限。政策 Excel 导入需要 `policy:write` 权限。消息详情接口只允许读取当前用户自己会话下的消息。

错误格式：

```json
{
  "error": {
    "code": "validation_error",
    "message": "错误说明",
    "requestId": "req_123"
  }
}
```

## 登录

### POST `/auth/login`

请求：

```json
{
  "account": "admin",
  "password": "admin123"
}
```

响应：

```json
{
  "token": "session-token",
  "user": {
    "id": "ObjectID",
    "account": "admin",
    "displayName": "张医生",
    "roles": ["系统管理员", "知识库管理员"],
    "permissions": ["chat:use", "knowledge:read", "knowledge:write", "system:read"]
  }
}
```

### GET `/auth/me`

返回当前登录用户。

### POST `/auth/logout`

退出当前会话。

## 系统设置

### GET `/system/model-config`

返回当前模型配置。API Key 只返回是否已配置和脱敏预览，不返回明文。

### PATCH `/system/model-config`

保存模型配置到 MongoDB。`deepSeekAPIKey` 或 `qwenEmbeddingAPIKey` 留空时保留已保存密钥。

```json
{
  "deepSeekBaseUrl": "https://ark.cn-beijing.volces.com/api/v3",
  "deepSeekAPIKey": "sk-...",
  "deepSeekChatModel": "deepseek-v4-flash-260425",
  "qwenEmbeddingBaseUrl": "https://ark.cn-beijing.volces.com/api/v3",
  "qwenEmbeddingAPIKey": "sk-...",
  "qwenEmbeddingModel": "doubao-embedding-vision-251215",
  "qwenEmbeddingDimension": 2048
}
```

## 知识库

### GET `/knowledge-bases`

查询参数：

- `scenario`
- `tag`
- `department`
- `keyword`
- `page`
- `size`

响应：

```json
{
  "items": [],
  "total": 5,
  "page": 1,
  "size": 10
}
```

### POST `/knowledge-bases`

创建知识库。

```json
{
  "name": "糖尿病诊疗知识库",
  "description": "诊疗规范和指南",
  "scenario": "临床诊疗",
  "tags": ["诊疗", "指南"],
  "department": "医务部"
}
```

### PATCH `/knowledge-bases/{id}`

更新知识库元数据或状态。

### GET `/knowledge-bases/{id}/documents`

获取知识库文档列表。

### POST `/knowledge-bases/{id}/documents`

上传文档，使用 `multipart/form-data`：

```text
file=<上传文件>
```

支持：PDF、Word(.docx)、Excel(.xlsx/.xls)、Markdown、TXT、CSV。单个文件不超过 15MB，原始文件会保存到 MongoDB 并进入预处理、切分、向量化流程。

### GET `/knowledge-bases/{id}/documents/{documentId}`

查看文档预处理后的文本内容。PDF/Word/Excel 会按入库预处理逻辑提取正文。

### GET `/knowledge-bases/{id}/documents/{documentId}/chunks`

查看文档已入库的分片内容，按 `chunkIndex` 升序返回。

### GET `/knowledge-bases/{id}/documents/{documentId}/download`

下载 MongoDB 中保存的原始文件。

## 入库任务

### GET `/ingestion-jobs/{id}`

查看文档解析、切分、向量化和索引状态。

### POST `/ingestion-jobs/{id}:retry`

重试失败的入库任务。

## 复盘笔记

复盘笔记保存在独立的 `review_notes` / `review_note_exports` 集合中，不进入知识库文档、chunks、ingestion jobs 或 Qdrant 向量索引。

### GET `/review-notes`

分页查询复盘笔记。支持参数：

- `page`：默认 1
- `size`：默认 10，最大 100

响应：

```json
{
  "total": 20,
  "page": 1,
  "size": 10,
  "items": [
    {
      "id": "ObjectID",
      "content": "复盘内容",
      "exported": false,
      "createdAt": "2026-06-17T09:00:00Z"
    }
  ]
}
```

### GET `/review-notes/counts`

返回累计记录数和未导出记录数。

```json
{
  "total": 20,
  "unexported": 15
}
```

### POST `/review-notes`

创建复盘笔记。

```json
{
  "content": "记录一次可复用的判断、流程口径或风险提醒"
}
```

### POST `/review-notes:export`

将选中的复盘笔记生成 Markdown 附件并直接下载。

请求：

```json
{
  "noteIds": ["ObjectID"]
}
```

响应头：

```text
Content-Type: text/markdown; charset=utf-8
Content-Disposition: attachment; filename="..."; filename*=UTF-8''...
```

导出成功后，包含的记录会标记为已导出并关联导出批次；导出批次保存生成文档快照，后续删除原始记录不影响历史下载。系统最多保留最近 5 条导出文档。

### DELETE `/review-notes/{id}`

删除单条复盘记录。需要 `review_notes:write` 权限。

### GET `/review-notes/exports`

查看历史导出批次，支持 `limit` 参数，最大返回最近 5 条。前端用它提供“再次下载”入口。

### GET `/review-notes/exports/{id}/download`

重复下载指定历史导出批次，返回同样的 Markdown 附件响应头和内容格式。

## 政策文件库

政策文件库保存在独立的 `policy_documents` / `policy_import_batches` 集合中，不进入知识库入库或向量检索链路。

固定分类：

```text
国家医学中心、科技创新、医疗服务、医保医药、数智治理、改革监管、其他
```

### GET `/policies/categories`

返回固定分类列表。

### GET `/policies`

查询政策列表。支持参数：

- `category`
- `date`：按月份或日期前缀筛选，例如 `2026-06`
- `keyword`
- `page`：默认 1
- `size`：默认 10，最大 100

响应：

```json
{
  "total": 42,
  "page": 1,
  "size": 10,
  "items": [
    {
      "id": "ObjectID",
      "title": "政策标题",
      "summary": "政策摘要",
      "interpretation": "政策解读",
      "date": "2026-06-08",
      "category": "医疗服务"
    }
  ],
  "facets": {
    "categories": [{ "value": "医疗服务", "count": 3 }],
    "dates": [{ "value": "2026-06", "count": 2 }]
  }
}
```

### DELETE `/policies/{id}`

删除单条政策记录。需要 `policy:write` 权限。

### GET `/policies/import-template`

下载政策文件库导入模板。响应为 `.xlsx` 附件，模板表头为：

```text
标题、摘要、解读、日期、分类标签
```

### POST `/policies:import`

导入政策 Excel，使用 `multipart/form-data`：

```text
file=<政策库.xlsx>
```

推荐使用导入模板。Excel 至少需要包含标题、摘要、日期、分类字段；支持 `分类标签` 作为分类字段，支持可选 `解读` 字段。分类必须属于固定七类。

## 对话

### GET `/conversations`

查询当前用户会话列表。支持 `keyword` 参数。

### POST `/conversations`

```json
{
  "title": "糖尿病治疗规范咨询",
  "knowledgeBaseIds": ["ObjectID"]
}
```

### GET `/conversations/{id}/messages`

获取消息历史。

### POST `/conversations/{id}/messages:stream`

流式问答接口，详见 [streaming-chat.md](streaming-chat.md)。

### GET `/messages/{id}/details`

查看回复详情，包含：

- 回复内容
- 引用来源
- 模型名称
- 提示词上下文
- 响应耗时
