import { Navigate, Route, Routes } from 'react-router-dom';
import { useAuth } from './contexts/AuthContext.jsx';
import Shell from './layouts/Shell.jsx';
import LoginPage from './pages/LoginPage.jsx';
import ChatPage from './pages/ChatPage.jsx';
import KnowledgePage from './pages/KnowledgePage.jsx';
import ReviewNotesPage from './pages/ReviewNotesPage.jsx';
import PolicyLibraryPage from './pages/PolicyLibraryPage.jsx';

function ProtectedRoute({ children }) {
  const { token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  return children;
}

export default function App() {
  const { token } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={token ? <Navigate to="/chat" replace /> : <LoginPage />} />
      <Route
        path="/chat"
        element={(
          <ProtectedRoute>
            <Shell>
              <ChatPage />
            </Shell>
          </ProtectedRoute>
        )}
      />
      <Route
        path="/knowledge"
        element={(
          <ProtectedRoute>
            <Shell>
              <KnowledgePage />
            </Shell>
          </ProtectedRoute>
        )}
      />
      <Route
        path="/review-notes"
        element={(
          <ProtectedRoute>
            <Shell>
              <ReviewNotesPage />
            </Shell>
          </ProtectedRoute>
        )}
      />
      <Route
        path="/policies"
        element={(
          <ProtectedRoute>
            <Shell>
              <PolicyLibraryPage />
            </Shell>
          </ProtectedRoute>
        )}
      />
      <Route path="*" element={<Navigate to={token ? '/chat' : '/login'} replace />} />
    </Routes>
  );
}
