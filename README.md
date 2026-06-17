# 医院行政智策平台

这是一个基于 RAG 技术的医院知识库对话系统，包含：

- React + React Router 前端
- Go + chi 后端
- MongoDB 数据维护
- Qdrant 向量检索
- DeepSeek 流式对话
- Qwen3-Embedding 文档和查询向量化

## 快速启动

首次准备：

```bash
sh scripts/setup.sh
```

一键启动：

```bash
sh scripts/start-all.sh
```

Docker 安装、MongoDB/Qdrant 容器查看说明：

```text
docs/runbooks/docker和数据库查看.md
```

手动清空业务数据库并保留一个 admin 账号：

```bash
sh scripts/reset-database.sh
```

该脚本会清空知识库、会话、复盘记录、政策文件库等业务集合，并重建 Qdrant collection。

## 默认账号

```text
账号：admin
密码：admin123
```

## 目录

- `frontend/`：React 前端界面，按 `api/`、`contexts/`、`layouts/`、`pages/`、`features/`、`components/` 拆分
- `backend/`：Go 后端服务，HTTP 层使用 chi，Mongo store 按领域拆分
- `scripts/`：本地 setup、启动、构建和测试脚本
- `backend/scripts/mongo-init.js`：MongoDB 集合和索引初始化脚本
- `backend/scripts/qdrant-init.sh`：Qdrant collection 初始化脚本
- `docs/api/`：接口对接文档
- `docs/runbooks/`：启动构建说明
- `openspec/`：OpenSpec 需求和任务记录
