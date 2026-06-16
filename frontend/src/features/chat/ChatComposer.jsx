export default function ChatComposer({ input, setInput, streaming, onSend }) {
  return (
    <div className="composer">
      <textarea
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            onSend();
          }
        }}
        placeholder="输入消息，Enter 发送，Shift + Enter 换行"
      />
      <div className="composer-tools">📎 ▣ ⏱</div>
      <button className="send" onClick={onSend} disabled={streaming}>➤</button>
    </div>
  );
}
