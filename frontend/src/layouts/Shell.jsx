import { useMemo, useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { Button, Layout, Space } from 'antd';
import { DatabaseOutlined, FileSearchOutlined, FormOutlined, LogoutOutlined, MessageOutlined, SettingOutlined, UserOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext.jsx';
import ModelSettingsModal from '../features/settings/ModelSettingsModal.jsx';

const { Header, Content, Sider } = Layout;

export default function Shell({ children }) {
  const { user, logout } = useAuth();
  const location = useLocation();
  const [settingsOpen, setSettingsOpen] = useState(false);

  const navItems = useMemo(() => [
    { to: '/chat', label: '智能问答', icon: <MessageOutlined /> },
    { to: '/policies', label: '政策文件库', icon: <FileSearchOutlined /> },
    { to: '/review-notes', label: '复盘笔记', icon: <FormOutlined /> },
    { to: '/knowledge', label: '知识库管理', icon: <DatabaseOutlined /> },
  ], []);

  return (
    <Layout className="app-shell">
      <Header className="app-header">
        <div className="top-brand">
          <span className="brand-mark">研</span>
          <div>
            <b>医院行政智策平台v1.0</b>
          </div>
        </div>
        <Space size={8} className="top-actions">
          <Button icon={<SettingOutlined />} onClick={() => setSettingsOpen(true)}>
            系统设置
          </Button>
          <Button icon={<LogoutOutlined />} onClick={logout}>
            退出
          </Button>
          <Button icon={<UserOutlined />} disabled>
            {user?.displayName || user?.account || '管理员'}
          </Button>
        </Space>
      </Header>
      <Layout className="app-body">
        <Sider width={176} className="app-nav-sider">
          <nav className="side-nav">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={location.pathname === item.to ? 'side-nav-item active' : 'side-nav-item'}
              >
                {item.icon}
                <span>{item.label}</span>
              </NavLink>
            ))}
          </nav>
        </Sider>
        <Content className="app-content">{children}</Content>
      </Layout>
      <ModelSettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </Layout>
  );
}
