import WelcomeMessage from './WelcomeMessage.jsx';

export default function MessageList({ messages, onShowDetail }) {
  return (
    <div className="messages">
      {messages.length === 0 && <WelcomeMessage />}
      {messages.map((msg) => (
        <article key={msg.id} className={msg.role === 'user' ? 'message user' : 'message assistant'}>
          {msg.role === 'assistant' && <div className="bot-icon">中</div>}
          <div className="bubble">
            <p>{msg.content || (msg.status === 'streaming' ? '正在生成...' : '')}</p>
            {msg.citations?.length > 0 && (
              <div className="citations">
                <b>引用来源</b>
                {msg.citations.slice(0, 3).map((cite, idx) => <span key={`${cite.chunkId}-${idx}`}>{idx + 1}. {cite.title}</span>)}
              </div>
            )}
            {msg.role === 'assistant' && (
              <div className="message-actions">
                <button>↻</button>
                <button>⧉</button>
                <button onClick={() => onShowDetail(msg)}>▣</button>
              </div>
            )}
          </div>
        </article>
      ))}
    </div>
  );
}
