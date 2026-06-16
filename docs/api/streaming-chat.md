# 流式对话接口

接口：

```text
POST /api/v1/conversations/{conversationId}/messages:stream
Accept: text/event-stream
Authorization: Bearer <token>
```

请求体：

```json
{
  "content": "请问2型糖尿病的最新诊疗规范是什么？",
  "knowledgeBaseIds": ["知识库ObjectID"]
}
```

## SSE 事件

### `message.started`

表示后端已创建 AI 回复消息。

```json
{
  "messageId": "ObjectID",
  "conversationId": "ObjectID"
}
```

### `retrieval.sources`

返回本次 RAG 检索引用来源。

```json
{
  "sources": [
    {
      "chunkId": "ObjectID",
      "documentId": "ObjectID",
      "knowledgeBaseId": "ObjectID",
      "title": "2型糖尿病防治指南2023.pdf",
      "snippet": "2型糖尿病治疗应包括生活方式干预...",
      "score": 0.82
    }
  ]
}
```

### `message.delta`

模型增量输出。

```json
{
  "text": "根据指南，"
}
```

### `message.completed`

流式输出完成。

```json
{
  "messageId": "ObjectID",
  "durationMs": 2350,
  "citationCount": 3
}
```

### `message.error`

生成失败。

```json
{
  "code": "model_error",
  "message": "DeepSeek 返回状态码 500",
  "requestId": "req_123"
}
```

## 前端处理要求

- 收到 `message.delta` 后立即追加到当前 AI 回复气泡。
- 收到 `retrieval.sources` 后展示引用来源。
- 收到 `message.completed` 后把消息状态改为完成。
- 收到 `message.error` 后展示可恢复失败状态。
