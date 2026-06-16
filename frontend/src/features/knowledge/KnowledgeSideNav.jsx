export default function KnowledgeSideNav() {
  return (
    <aside className="side-nav">
      <button>💬 对话管理</button>
      <button className="open">▣ 知识库管理</button>
      <button className="sub active">知识库列表</button>
      <button className="sub">文档上传</button>
      <button className="sub">场景管理</button>
      <button className="sub">标签管理</button>
      <button className="sub">部门管理</button>
      <button>⚙ 系统设置</button>
    </aside>
  );
}
