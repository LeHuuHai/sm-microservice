import React from 'react';
import { Outlet, Navigate } from 'react-router-dom';
import Sidebar from './Sidebar';
import { tokenService } from '../api/axios';

const Layout: React.FC = () => {
  const isAuthenticated = !!tokenService.getAccessToken();

  if (!isAuthenticated) return <Navigate to="/login" replace />;

  return (
    <div className="app-container">
      <Sidebar />
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
};

export default Layout;
