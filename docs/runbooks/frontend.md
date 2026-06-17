# 前端启动和构建说明

前端目录：

```text
frontend/
```

技术栈：

- JavaScript
- React
- Vite
- React Router
- 原生 CSS

## 环境变量

复制模板：

```bash
cd frontend
cp .env.example .env
```

默认内容：

```text
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 启动开发服务

```bash
cd frontend
npm install
npm run dev
```

访问：

```text
http://localhost:5173
```

## 构建

```bash
cd frontend
npm run build
```

构建产物在：

```text
frontend/dist/
```

## 页面说明

- 登录页：严格按照 `design.png` 的左侧品牌区和中间登录卡片实现。
- 对话页：包含会话列表、消息区、输入框、引用来源、回复详情抽屉。
- 知识库页：包含侧边导航、筛选区、知识库表格、右侧文档上传面板。
- 复盘笔记页：左侧导航进入，包含笔记输入、未导出/累计计数、服务端分页记录列表；记录列表支持勾选、删除，用户可将选中的记录生成 Markdown 并下载。导出记录保留最近 5 条生成文档且支持再次下载。对话页消息操作可把消息内容带入复盘笔记草稿，用户编辑后再提交入库。
- 政策文件库页：左侧固定七个分类并显示聚合数量，右侧分页展示政策标题、摘要、解读、日期和分类标签；支持下载导入模板，按日期聚合筛选，具备 `policy:write` 权限时可导入 `.xlsx` 并删除单条政策记录。

## 前端代码结构

```text
frontend/src/
  main.jsx                 # 应用挂载入口
  App.jsx                  # React Router 路由和登录保护
  api/client.js            # REST API 客户端和流式对话请求
  api/sse.js               # SSE 事件解析
  contexts/AuthContext.jsx # 登录态、用户信息、退出登录
  layouts/Shell.jsx        # 顶部品牌栏、左侧导航和认证后布局
  pages/                   # LoginPage、ChatPage、KnowledgePage、ReviewNotesPage、PolicyLibraryPage
  features/chat/           # 对话业务组件和 useChatWorkspace
  features/knowledge/      # 知识库业务组件和 useKnowledgeWorkspace
  features/reviewNotes/    # 复盘笔记逻辑和 useReviewNotesWorkspace
  features/policyLibrary/  # 政策分类、导入校验和 usePolicyLibraryWorkspace
  components/              # 通用组件
  utils/                   # 格式化工具
```

本次重构后前端不再把 API、登录态、页面和业务组件集中在 `main.jsx` 中，后续新增页面应优先放在 `pages/`，业务逻辑优先放在对应 `features/*/use*.js` hook 中。

## 默认账号

```text
账号：admin
密码：admin123
```
