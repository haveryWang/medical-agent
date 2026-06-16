import { Drawer, Layout, Typography } from 'antd';
import AnswerDetail from '../features/chat/AnswerDetail.jsx';
import ChatComposer from '../features/chat/ChatComposer.jsx';
import ConversationList from '../features/chat/ConversationList.jsx';
import MessageList from '../features/chat/MessageList.jsx';
import { useChatWorkspace } from '../features/chat/useChatWorkspace.js';

const { Sider, Content } = Layout;

export default function ChatPage() {
  const chat = useChatWorkspace();

  return (
    <Layout className="chat-layout">
      <Sider width={292} className="workspace-sider">
        <ConversationList
          active={chat.active}
          conversations={chat.conversations}
          knowledgeBases={chat.knowledgeBases}
          keyword={chat.searchKeyword}
          loading={chat.loadingConversations}
          onDelete={chat.deleteConversation}
          onKnowledgeBaseChange={chat.updateKnowledgeBaseScope}
          onKeywordChange={chat.setSearchKeyword}
          onNew={chat.newConversation}
          onOpen={chat.openConversation}
          onRename={chat.renameConversation}
          onSearch={() => chat.loadConversations(chat.searchKeyword)}
          selectedKnowledgeBaseIds={chat.selectedKnowledgeBaseIds}
          serviceStatus={chat.serviceStatus}
        />
      </Sider>
      <Content className="message-pane">
        <header className="pane-header">
          <div>
            <Typography.Title level={4}>{chat.active?.title || '对话管理'}</Typography.Title>
            <Typography.Text type="secondary">{chat.active ? chat.paneHint : '创建会话后开始提问'}</Typography.Text>
          </div>
        </header>
        <MessageList
          loading={chat.loadingMessages}
          messages={chat.messages}
          onRegenerate={chat.regenerate}
          onShowDetail={chat.showDetail}
        />
        <ChatComposer input={chat.input} setInput={chat.setInput} streaming={chat.streaming} onSend={chat.send} />
      </Content>
      <Drawer
        title="回复详情"
        size={820}
        open={chat.detailDrawerOpen}
        onClose={chat.closeDetail}
      >
        <AnswerDetail detail={chat.detail} loading={chat.detailLoading} />
      </Drawer>
    </Layout>
  );
}
