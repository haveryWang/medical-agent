const fallbackTitles = ['糖尿病治疗规范咨询', '高血压用药建议', '冠心病治疗方案讨论'];

export default function ConversationList({ active, conversations, onNew, onOpen }) {
  return (
    <aside className="conversation-pane">
      <button className="new-chat" onClick={onNew}>＋ 新建对话</button>
      <div className="search">🔍 搜索会话</div>
      <div className="conversation-list">
        {conversations.map((conv, index) => (
          <button key={conv.id} onClick={() => onOpen(conv)} className={active?.id === conv.id ? 'conv active' : 'conv'}>
            <strong>{conv.title || fallbackTitles[index % fallbackTitles.length]}</strong>
            <small>{index === 0 ? '14:32' : index === 1 ? '10:15' : '昨天'}</small>
            <span>请问相关诊疗规范和用药建议...</span>
          </button>
        ))}
      </div>
      <button className="trash">🗑 回收站</button>
    </aside>
  );
}
