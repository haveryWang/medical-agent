import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext.jsx';

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [account, setAccount] = useState('admin');
  const [password, setPassword] = useState('admin123');
  const [remember, setRemember] = useState(true);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function submit(event) {
    event.preventDefault();
    setLoading(true);
    setError('');
    try {
      await login(account, password);
      navigate('/chat', { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="login-screen">
      <section className="login-brand">
        <div className="brand-row">
          <div className="shield">研</div>
          <div>
            <h1>医院知识库管理平台</h1>
            <p>构建专业知识体系 · 提升医疗服务质量</p>
          </div>
        </div>
        <div className="hospital-illustration">
          <div className="building b1" />
          <div className="building b2" />
          <div className="building b3" />
          <div className="trees" />
        </div>
      </section>
      <form className="login-card" onSubmit={submit}>
        <h2>用户登录</h2>
        <label className="input-line">
          <span>👤</span>
          <input value={account} onChange={(e) => setAccount(e.target.value)} placeholder="请输入账号" />
        </label>
        <label className="input-line">
          <span>🔒</span>
          <input value={password} type="password" onChange={(e) => setPassword(e.target.value)} placeholder="请输入密码" />
          <small>∞</small>
        </label>
        <label className="remember">
          <input type="checkbox" checked={remember} onChange={(e) => setRemember(e.target.checked)} />
          记住我
        </label>
        {error && <div className="form-error">{error}</div>}
        <button className="primary" disabled={loading}>{loading ? '登录中...' : '登录'}</button>
        <footer>© 2024 医院知识库管理平台 版权所有</footer>
      </form>
    </main>
  );
}
