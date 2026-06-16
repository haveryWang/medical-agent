import { useMemo, useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { Avatar, Button, Layout, Space, Typography } from 'antd';
import { DatabaseOutlined, LogoutOutlined, MessageOutlined, SettingOutlined, UserOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext.jsx';
import ModelSettingsModal from '../features/settings/ModelSettingsModal.jsx';

const { Header, Content } = Layout;

export default function Shell({ children }) {
  const { user, logout } = useAuth();
  const location = useLocation();
  const [settingsOpen, setSettingsOpen] = useState(false);

  const navItems = useMemo(() => [
    { to: '/chat', label: '对话管理', icon: <MessageOutlined /> },
    { to: '/knowledge', label: '知识库管理', icon: <DatabaseOutlined /> },
  ], []);

  return (
    <Layout className="app-shell">
      <Header className="app-header">
        <div className="top-brand">
          <span className="brand-mark">研</span>
          <div>
            <b>医院知识库管理平台</b>
          </div>
        </div>
        <Space size={8} className="top-actions">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} className={location.pathname === item.to ? 'nav-pill active' : 'nav-pill'}>
              {item.icon}
              <span>{item.label}</span>
            </NavLink>
          ))}
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
      <Content className="app-content">{children}</Content>
      <ModelSettingsModal open={settingsOpen} onClose={() => setSettingsOpen(false)} />
    </Layout>
  );
}
