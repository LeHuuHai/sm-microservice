import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Server as ServerIcon, LogOut } from 'lucide-react';
import api, { tokenService } from '../api/axios';

const Sidebar: React.FC = () => {
  const navigate = useNavigate();

  const handleLogout = async () => {
    try { await api.post('/auth/logout'); } 
    catch (e) { console.error('Logout failed on backend'); } 
    finally {
      tokenService.clearTokens();
      navigate('/login');
    }
  };

  return (
    <div className="sidebar">
      <div className="sidebar-logo">
        <ServerIcon size={24} color="var(--accent-primary)" />
        <span>SM System</span>
      </div>

      <nav className="sidebar-nav" style={{ flex: 1 }}>
        <NavLink to="/" className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'} end>
          <LayoutDashboard size={20} />
          Dashboard
        </NavLink>
        <NavLink to="/servers" className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}>
          <ServerIcon size={20} />
          Servers
        </NavLink>
      </nav>

      <div style={{ marginTop: 'auto', paddingTop: '1rem', borderTop: '1px solid var(--border-light)' }}>
        <button onClick={handleLogout} className="nav-item" style={{ width: '100%', background: 'transparent', border: 'none', cursor: 'pointer', textAlign: 'left' }}>
          <LogOut size={20} color="var(--accent-danger)" />
          <span style={{ color: 'var(--accent-danger)' }}>Logout</span>
        </button>
      </div>
    </div>
  );
};

export default Sidebar;
