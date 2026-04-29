import React from 'react';
import { useAuthStore } from '../../store/authStore';

export const Header: React.FC = () => {
  const { user, isAuthenticated, isGuest, logout } = useAuthStore();

  return (
    <header className="bg-white shadow-sm border-b">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <h1 className="text-xl font-bold text-primary-500">
              AI 面试系统
            </h1>
          </div>

          <div className="flex items-center gap-4">
            {isAuthenticated && (
              <>
                <span className="text-sm text-gray-600">
                  {isGuest ? '游客模式' : user?.username || user?.email}
                </span>
                <button
                  onClick={logout}
                  className="text-sm text-gray-600 hover:text-gray-900 transition-colors"
                >
                  退出
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};
