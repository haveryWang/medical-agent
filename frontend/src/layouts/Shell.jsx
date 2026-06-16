import { NavLink } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext.jsx';

export default function Shell({ children }) {
  const { user, logout } = useAuth();

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="top-brand">
          <span className="mini-shield">研</span>
          <strong>医院知识库管理平台</strong>
        </div>
        <nav className="top-actions">
          <NavLink className={({ isActive }) => (isActive ? 'tab active' : 'tab')} to="/chat">对话管理</NavLink>
          <NavLink className={({ isActive }) => (isActive ? 'tab active' : 'tab')} to="/knowledge">知识库管理</NavLink>
          <span className="avatar">👤</span>
          <span>{user?.displayName || '张医生'}</span>
          <button className="ghost" onClick={logout}>退出</button>
        </nav>
      </header>
      {children}
    </main>
  );
}
