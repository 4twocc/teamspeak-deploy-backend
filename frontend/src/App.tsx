// src/App.tsx
import { BrowserRouter, Route, Routes, Navigate } from 'react-router-dom';
import Login from '@/pages/Login';
import Dashboard from '@/pages/Dashboard';
import Instances from '@/pages/Instances';
import Users from '@/pages/User';
import { Layout } from '@/components/base/Layout';
import { ErrorBoundary } from 'react-error-boundary';

function ErrorFallback({ error }: { error: Error }) {
  return (
    <div className="p-8 text-red-500">
      <h2 className="text-xl font-bold mb-4">出错了</h2>
      <pre className="whitespace-pre-wrap">{error.message}</pre>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          element={
            <ErrorBoundary FallbackComponent={ErrorFallback}>
              <Layout />
            </ErrorBoundary>
          }
        >
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/instances" element={<Instances />} />
          <Route path="/users" element={<Users />} />
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
