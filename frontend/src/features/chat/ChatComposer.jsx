import { Button, Input } from 'antd';
import { SendOutlined } from '@ant-design/icons';

export default function ChatComposer({ input, setInput, streaming, onSend }) {
  return (
    <div className="composer">
      <Input.TextArea
        value={input}
        onChange={(event) => setInput(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            onSend();
          }
        }}
        autoSize={{ minRows: 2, maxRows: 5 }}
        placeholder="输入消息，Enter 发送，Shift + Enter 换行"
      />
      <Button type="primary" icon={<SendOutlined />} loading={streaming} disabled={!input.trim()} onClick={() => onSend()}>
        发送
      </Button>
    </div>
  );
}
