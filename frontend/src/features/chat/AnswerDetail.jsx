export default function AnswerDetail({ detail, setDetail }) {
  return (
    <aside className="detail-pane">
      <div className="detail-head">
        <h3>回复详情</h3>
        <button onClick={() => setDetail(null)}>×</button>
      </div>
      {detail ? (
        <>
          <dl>
            <dt>消息ID</dt><dd>{detail.id}</dd>
            <dt>对话耗时</dt><dd>{detail.durationMs ? `${(detail.durationMs / 1000).toFixed(2)}s` : '2.35s'}</dd>
          </dl>
          <h4>知识引用（{detail.citations?.length || 0}条）</h4>
          <ol className="reference-list">
            {(detail.citations || []).map((cite, idx) => <li key={`${cite.chunkId}-${idx}`}><span>{cite.title}</span><em>{(cite.score * 100).toFixed(0)}%</em></li>)}
          </ol>
          <h4>加载优化策略</h4>
          <p className="muted">混合检索（向量检索 + 关键词检索）TopK: 5</p>
          <h4>发送给模型的原文</h4>
          <pre>{detail.promptContext || '选择一条 AI 回复查看完整上下文。'}</pre>
        </>
      ) : (
        <p className="empty-detail">点击回复右侧详情按钮查看引用、耗时和提示词上下文。</p>
      )}
    </aside>
  );
}
