import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App as AntApp, ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import App from './App.jsx';
import { AuthProvider } from './contexts/AuthContext.jsx';
import 'antd/dist/reset.css';
import './styles.css';

createRoot(document.getElementById('root')).render(
  <ConfigProvider
    locale={zhCN}
    theme={{
      token: {
        colorPrimary: '#1677ff',
        colorInfo: '#1677ff',
        colorSuccess: '#12966d',
        colorWarning: '#d48806',
        colorError: '#cf1322',
        borderRadius: 6,
        fontFamily: '"PingFang SC", "Microsoft YaHei", "Noto Sans SC", sans-serif',
      },
      components: {
        Button: { borderRadius: 6 },
        Card: { borderRadiusLG: 8 },
        Layout: { bodyBg: '#f3f6fb', headerBg: '#ffffff', siderBg: '#ffffff' },
      },
    }}
  >
    <AntApp>
      <BrowserRouter>
        <AuthProvider>
          <App />
        </AuthProvider>
      </BrowserRouter>
    </AntApp>
  </ConfigProvider>,
);
