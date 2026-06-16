export default function KnowledgeTable({ items, selected, onChoose }) {
  return (
    <div className="table-card">
      <table>
        <thead>
          <tr><th>知识库名称</th><th>场景</th><th>标签</th><th>所属部门</th><th>文档数量</th><th>构建状态</th><th>更新时间</th><th>操作</th></tr>
        </thead>
        <tbody>
          {items.map((kb) => (
            <tr key={kb.id} onClick={() => onChoose(kb)} className={selected?.id === kb.id ? 'selected-row' : ''}>
              <td><b>{kb.name}</b><span>{kb.description}</span></td>
              <td>{kb.scenario}</td>
              <td>{kb.tags?.join('、')}</td>
              <td>{kb.department}</td>
              <td>{kb.documentCount}</td>
              <td><em className={kb.buildStatus === 'building' ? 'status building' : 'status'}>{kb.buildStatus === 'building' ? '构建中' : '已完成'}</em></td>
              <td>{new Date(kb.updatedAt).toLocaleDateString()}</td>
              <td><a>查看</a><a>编辑</a><a>更多</a></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
