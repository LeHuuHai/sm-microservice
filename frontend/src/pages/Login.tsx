import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Server, Lock, User, LogIn } from 'lucide-react';
import api, { tokenService } from '../api/axios';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const response = await api.post('/auth/login', { username, password });
      
      if (response.data?.access_token && response.data?.refresh_token) {
        tokenService.setTokens(response.data.access_token, response.data.refresh_token);
        navigate('/');
      } else {
        setError('Invalid response from server.');
      }
    } catch (err: any) {
      if (err.response && err.response.status === 401) {
        setError('Invalid username or password.');
      } else {
        setError('Failed to connect to the authentication server.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', width: '100%' }}>
      <div className="glass-panel" style={{ width: '100%', maxWidth: '420px', padding: '2.5rem', animation: 'modal-slide-up 0.5s ease-out' }}>
        <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
          <div style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: '64px', height: '64px', borderRadius: '50%', background: 'rgba(59, 130, 246, 0.1)', color: 'var(--accent-primary)', marginBottom: '1rem', boxShadow: 'var(--shadow-glow)' }}>
            <Server size={32} />
          </div>
          <h1 style={{ fontSize: '1.5rem', fontWeight: 600, marginBottom: '0.5rem' }}>Server Management</h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Enter your credentials to access the dashboard</p>
        </div>

        {error && (
          <div style={{ background: 'rgba(239, 68, 68, 0.1)', border: '1px solid rgba(239, 68, 68, 0.2)', color: 'var(--accent-danger)', padding: '0.75rem 1rem', borderRadius: 'var(--radius-sm)', marginBottom: '1.5rem', fontSize: '0.875rem', textAlign: 'center' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleLogin}>
          <div className="input-group">
            <label className="input-label">Username</label>
            <div style={{ position: 'relative' }}>
              <div style={{ position: 'absolute', top: '50%', left: '1rem', transform: 'translateY(-50%)', color: 'var(--text-muted)' }}><User size={18} /></div>
              <input type="text" className="input-field" placeholder="admin" value={username} onChange={(e) => setUsername(e.target.value)} style={{ paddingLeft: '2.75rem' }} required />
            </div>
          </div>
          <div className="input-group">
            <label className="input-label">Password</label>
            <div style={{ position: 'relative' }}>
              <div style={{ position: 'absolute', top: '50%', left: '1rem', transform: 'translateY(-50%)', color: 'var(--text-muted)' }}><Lock size={18} /></div>
              <input type="password" className="input-field" placeholder="••••••••" value={password} onChange={(e) => setPassword(e.target.value)} style={{ paddingLeft: '2.75rem' }} required />
            </div>
          </div>
          <button type="submit" className="btn btn-primary" style={{ width: '100%', marginTop: '1rem' }} disabled={isLoading}>
            {isLoading ? 'Authenticating...' : <><LogIn size={18} />Sign In</>}
          </button>
        </form>
      </div>
    </div>
  );
};

export default Login;
