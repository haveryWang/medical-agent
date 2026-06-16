import { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { createApiClient } from '../api/client.js';

const TOKEN_KEY = 'medical_agent_token';
const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(localStorage.getItem(TOKEN_KEY) || '');
  const [user, setUser] = useState(null);
  const tokenRef = useRef(token);

  useEffect(() => {
    tokenRef.current = token;
  }, [token]);

  const clearSession = () => {
    localStorage.removeItem(TOKEN_KEY);
    setToken('');
    setUser(null);
  };

  const api = useMemo(() => createApiClient({
    getToken: () => tokenRef.current,
    onUnauthorized: clearSession,
  }), []);

  useEffect(() => {
    if (!token) return;
    api.me().then((res) => setUser(res.user)).catch(clearSession);
  }, [token, api]);

  async function login(account, password) {
    const res = await api.login(account, password);
    localStorage.setItem(TOKEN_KEY, res.token);
    setToken(res.token);
    setUser(res.user);
  }

  async function logout() {
    if (tokenRef.current) {
      await api.request('/auth/logout', { method: 'POST' }).catch(() => null);
    }
    clearSession();
  }

  const value = useMemo(() => ({ api, token, user, login, logout }), [api, token, user]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used inside AuthProvider');
  return context;
}
