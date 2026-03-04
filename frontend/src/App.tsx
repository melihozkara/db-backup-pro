import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ToastProvider } from './components/Toast';
import { ConfirmProvider } from './components/ConfirmDialog';
import { TooltipProvider } from './components/ui/tooltip';
import ErrorBoundary from './components/ErrorBoundary';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Databases from './pages/Databases';
import History from './pages/History';
import Settings from './pages/Settings';

function App() {
  return (
    <ErrorBoundary>
      <ToastProvider>
        <ConfirmProvider>
          <TooltipProvider>
            <BrowserRouter>
              <Routes>
                <Route path="/" element={<Layout />}>
                  <Route index element={<Dashboard />} />
                  <Route path="databases" element={<Databases />} />
                  <Route path="history" element={<History />} />
                  <Route path="settings" element={<Settings />} />
                </Route>
              </Routes>
            </BrowserRouter>
          </TooltipProvider>
        </ConfirmProvider>
      </ToastProvider>
    </ErrorBoundary>
  );
}

export default App;
