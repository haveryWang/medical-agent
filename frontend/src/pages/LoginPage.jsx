import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Alert, Button, Card, Checkbox, Form, Input, Typography } from 'antd';
import { LockOutlined, SafetyCertificateOutlined, UserOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext.jsx';

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function submit(values) {
    setLoading(true);
    setError('');
    try {
      await login(values.account, values.password);
      navigate('/chat', { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="login-screen">
      <section className="login-identity">
        <div className="brand-lockup">
          <span className="brand-mark large">研</span>
          <div>
            <Typography.Title level={1}>医院行政智策平台v1.0</Typography.Title>
            <Typography.Paragraph>面向临床、药学、医保和护理的院内知识问答工作台</Typography.Paragraph>
          </div>
        </div>
        <div className="login-metrics">
          <div><b>RAG</b><span>知识检索增强</span></div>
          <div><b>Ark</b><span>火山引擎模型</span></div>
          <div><b>DB</b><span>配置与数据入库</span></div>
        </div>
      </section>
      <Card className="login-card" variant="borderless">
        <div className="login-card-title">
          <SafetyCertificateOutlined />
          <Typography.Title level={2}>用户登录</Typography.Title>
        </div>
        <Form
          layout="vertical"
          initialValues={{ account: 'admin', password: 'admin123', remember: true }}
          onFinish={submit}
          requiredMark={false}
        >
          <Form.Item name="account" label="账号" rules={[{ required: true, message: '请输入账号' }]}>
            <Input prefix={<UserOutlined />} placeholder="请输入账号" autoComplete="username" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" autoComplete="current-password" />
          </Form.Item>
          <Form.Item name="remember" valuePropName="checked" className="compact-form-item">
            <Checkbox>记住登录状态</Checkbox>
          </Form.Item>
          {error ? <Alert type="error" showIcon message={error} className="login-error" /> : null}
          <Button type="primary" htmlType="submit" block loading={loading} size="large">
            登录
          </Button>
        </Form>
      </Card>
    </main>
  );
}
