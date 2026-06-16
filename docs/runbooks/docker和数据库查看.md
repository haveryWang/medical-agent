# Docker 安装与数据库查看说明

本文档说明本项目本地开发所需的 Docker 安装、容器启动、MongoDB 数据查看、Qdrant 向量库查看和常见维护命令。

## 1. Docker 安装

### macOS

推荐安装 Docker Desktop：

1. 打开 Docker 官网下载页：`https://www.docker.com/products/docker-desktop/`
2. 下载 macOS 版本 Docker Desktop。
3. 安装后启动 Docker Desktop。
4. 等待右上角 Docker 图标显示为运行状态。
5. 在终端验证：

```bash
docker --version
docker compose version
```

### Windows

推荐安装 Docker Desktop：

1. 打开 Docker 官网下载页：`https://www.docker.com/products/docker-desktop/`
2. 下载 Windows 版本 Docker Desktop。
3. 安装时启用 WSL2 支持。
4. 安装完成后重启电脑。
5. 启动 Docker Desktop，等待运行完成。
6. 在 PowerShell 验证：

```powershell
docker --version
docker compose version
```

### Linux Ubuntu

可使用 Docker 官方 apt 源安装：

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

验证：

```bash
docker --version
docker compose version
sudo docker run hello-world
```

如需当前用户免 `sudo`：

```bash
sudo usermod -aG docker $USER
```

执行后退出并重新登录终端。
docker
## 2. 本项目容器说明

项目根目录的 `docker-compose.yml` 定义了两个服务：

```text
mongodb  -> MongoDB，端口 27017
qdrant   -> Qdrant 向量数据库，端口 6333 / 6334
```

启动：

```bash
docker compose up -d mongodb qdrant
```

查看状态：

```bash
docker compose ps
```

查看日志：

```bash
docker compose logs -f mongodb
docker compose logs -f qdrant
```

停止：

```bash
docker compose stop
```

停止并删除容器，但保留数据卷：

```bash
docker compose down
```

停止并删除容器和数据卷，会清空数据库：

```bash
docker compose down -v
```

## 3. MongoDB 初始化内容

MongoDB 容器第一次创建时会执行：

```text
backend/scripts/mongo-init.js
```

该脚本会创建以下集合和索引：

```text
users
sessions
roles
knowledge_bases
documents
chunks
ingestion_jobs
conversations
messages
audit_logs
model_configs
```

后端启动时也会自动检查索引，并在空库中插入演示数据：

```text
账号：admin
密码：admin123
```

## 4. 在 Docker 中查看 MongoDB 内容

进入 MongoDB 容器：

```bash
docker exec -it medical-agent-mongodb mongosh
```

切换数据库：

```javascript
use medical_agent
```

查看集合：

```javascript
show collections
```

查看用户：

```javascript
db.users.find().pretty()
```

查看知识库：

```javascript
db.knowledge_bases.find().pretty()
```

查看文档：

```javascript
db.documents.find().pretty()
```

查看切片：

```javascript
db.chunks.find().limit(5).pretty()
```

查看会话：

```javascript
db.conversations.find().pretty()
```

查看消息：

```javascript
db.messages.find().sort({ createdAt: -1 }).limit(10).pretty()
```

查看入库任务：

```javascript
db.ingestion_jobs.find().sort({ updatedAt: -1 }).pretty()
```

查看审计日志：

```javascript
db.audit_logs.find().sort({ createdAt: -1 }).limit(20).pretty()
```

统计各集合数量：

```javascript
db.users.countDocuments()
db.knowledge_bases.countDocuments()
db.documents.countDocuments()
db.chunks.countDocuments()
db.messages.countDocuments()
```

退出：

```javascript
exit
```

## 5. 使用 MongoDB Compass 图形化查看

如果想用图形界面查看：

1. 安装 MongoDB Compass。
2. 新建连接。
3. 连接地址填写：

```text
mongodb://localhost:27017
```

4. 打开数据库：

```text
medical_agent
```

然后可以在左侧查看集合和文档内容。

## 6. 查看 Qdrant 向量库内容

Qdrant 默认 HTTP 地址：

```text
http://localhost:6333
```

查看所有 collections：

```bash
curl http://localhost:6333/collections
```

查看本项目 collection：

```bash
curl http://localhost:6333/collections/medical_agent_chunks
```

查看 collection 点数量：

```bash
curl http://localhost:6333/collections/medical_agent_chunks
```

滚动查看向量点：

```bash
curl -X POST http://localhost:6333/collections/medical_agent_chunks/points/scroll \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 5,
    "with_payload": true,
    "with_vector": false
  }'
```

如果需要连向量一起查看：

```bash
curl -X POST http://localhost:6333/collections/medical_agent_chunks/points/scroll \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 1,
    "with_payload": true,
    "with_vector": true
  }'
```

## 7. 重新初始化数据库

如果需要清空所有数据并重新开始：

```bash
docker compose down -v
docker compose up -d mongodb qdrant
```

然后重新初始化 Qdrant：

```bash
cd backend
sh scripts/qdrant-init.sh
```

再启动后端，后端会重新插入演示用户和知识库数据。

## 8. 常见问题

### `docker compose` 命令不存在

检查 Docker Desktop 是否已启动，或 Docker Compose 插件是否安装：

```bash
docker compose version
```

老版本 Docker 可能使用：

```bash
docker-compose version
```

建议升级到支持 `docker compose` 的新版本。

### 端口 27017 被占用

说明本机已有 MongoDB 或其他容器占用端口。查看占用：

```bash
lsof -i :27017
```

可以停止本机 MongoDB，或修改 `docker-compose.yml` 的端口映射。

### 端口 6333 被占用

说明本机已有 Qdrant 或其他服务占用端口。查看占用：

```bash
lsof -i :6333
```

可以停止占用服务，或修改 `docker-compose.yml` 的端口映射。

### `mongosh` 命令不存在

不需要在本机安装 `mongosh`。直接进入容器执行：

```bash
docker exec -it medical-agent-mongodb mongosh
```

### MongoDB 看不到演示数据

演示用户和知识库由后端启动时插入。请确认后端至少启动过一次：

```bash
cd backend
go run ./cmd/server
```
