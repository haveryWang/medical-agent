# REST API 对接文档

Base URL:

```text
http://localhost:8080/api/v1
```

除登录接口外，所有接口都需要请求头：

```text
Authorization: Bearer <token>
```

创建/更新知识库、上传文档、重试入库任务需要当前用户拥有 `knowledge:write` 权限。消息详情接口只允许读取当前用户自己会话下的消息。

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
  "deepSeekChatModel": "DeepSeek-V4-flash",
  "qwenEmbeddingBaseUrl": "https://ark.cn-beijing.volces.com/api/v3",
  "qwenEmbeddingAPIKey": "sk-...",
  "qwenEmbeddingModel": "doubao-embedding-vision-251215",
  "qwenEmbeddingDimension": 1024
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

支持：PDF、Word、Excel、Markdown、文本文件。

## 入库任务

### GET `/ingestion-jobs/{id}`

查看文档解析、切分、向量化和索引状态。

### POST `/ingestion-jobs/{id}:retry`

重试失败的入库任务。

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
