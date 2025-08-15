// src/components/Layout.tsx
import { useState, useEffect } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import Button from '../ui/button';
import { Menu, X } from 'lucide-react';
import { NavLink } from './NavLink';
import { Toaster } from 'sonner';
import { cn } from '@/lib/utils';

export function Layout() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  // Close mobile menu when route changes
  useEffect(() => {
    setIsMenuOpen(false);
  }, [location]);

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile menu button */}
      <div className="lg:hidden bg-white shadow-sm">
        <div className="flex justify-between items-center px-4 h-16">
          <Button 
            className="hover:bg-gray-100 rounded-md p-2"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            {isMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </Button>
        </div>
      </div>

      {/* Sidebar */}
      <div className="lg:flex">
        <div 
          className={cn(
            "fixed inset-y-0 left-0 w-64 bg-white shadow-lg transform transition-transform duration-300 ease-in-out z-50",
            isMenuOpen ? "translate-x-0" : "-translate-x-full",
            "lg:translate-x-0 lg:static lg:inset-auto"
          )}
        >
          <nav className="h-full flex flex-col">
            <div className="p-4 border-b">
              <h1 className="text-xl font-bold">TS Manager</h1>
            </div>
            <div className="flex-1 p-4 space-y-1">
              <NavLink to="/dashboard">仪表盘</NavLink>
              <NavLink to="/instances">实例管理</NavLink>
              <NavLink to="/users">用户管理</NavLink>
            </div>
            <div className="p-4 border-t">
              <Button 
                className="w-full"
                onClick={handleLogout}
              >
                退出登录
              </Button>
            </div>
          </nav>
        </div>

        {/* Overlay for mobile */}
        {isMenuOpen && (
          <div 
            className="fixed inset-0 bg-black/50 z-40 lg:hidden"
            onClick={() => setIsMenuOpen(false)}
          />
        )}

        {/* Main content */}
        <main className="flex-1 p-4 lg:p-8 transition-all duration-300">
          {isLoading && (
            <div className="fixed inset-0 bg-white/50 flex items-center justify-center z-50">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-gray-900" />
            </div>
          )}
          <div className={cn(isLoading ? "opacity-50" : "opacity-100")}>
            <Outlet context={{ setIsLoading }} />
          </div>
        </main>
      </div>
      <Toaster position="top-center" />
    </div>
  );
}
