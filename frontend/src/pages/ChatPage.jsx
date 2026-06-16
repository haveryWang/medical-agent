import AnswerDetail from '../features/chat/AnswerDetail.jsx';
import ChatComposer from '../features/chat/ChatComposer.jsx';
import ConversationList from '../features/chat/ConversationList.jsx';
import MessageList from '../features/chat/MessageList.jsx';
import { useChatWorkspace } from '../features/chat/useChatWorkspace.js';

export default function ChatPage() {
  const chat = useChatWorkspace();

  return (
    <section className="chat-layout">
      <ConversationList
        active={chat.active}
        conversations={chat.conversations}
        onNew={chat.newConversation}
        onOpen={chat.openConversation}
      />
      <section className="message-pane">
        <div className="chat-title">{chat.active?.title || '糖尿病治疗规范咨询'}</div>
        <MessageList messages={chat.messages} onShowDetail={chat.showDetail} />
        <ChatComposer input={chat.input} setInput={chat.setInput} streaming={chat.streaming} onSend={chat.send} />
      </section>
      <AnswerDetail detail={chat.detail} setDetail={chat.setDetail} />
    </section>
  );
}
