import React, { useEffect, useState } from 'react';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import api from '../api/axios';

interface Server { id: string; name: string; ipv4: string; status: 'ONLINE' | 'OFFLINE' | 'UNKNOWN'; }
const COLORS = { ONLINE: '#10b981', OFFLINE: '#ef4444', UNKNOWN: '#6b7280' };

const Dashboard: React.FC = () => {
  const [servers, setServers] = useState<Server[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(12);
  const [total, setTotal] = useState(0);
  const [totalOnline, setTotalOnline] = useState(0);
  const [totalOffline, setTotalOffline] = useState(0);
  const [totalUnknown, setTotalUnknown] = useState(0);

  const fetchServers = async () => {
    try {
      const from = (page - 1) * pageSize;
      const to = page * pageSize;

      const response = await api.get('/monitor/livestatus', {
        params: { from, to }
      });
      const items = response.data.items || [];
      setTotal(response.data.total || 0);
      setTotalOnline(response.data.total_online || 0);
      setTotalOffline(response.data.total_offline || 0);
      setTotalUnknown(response.data.total_unknown || 0);

      const formatted = items.map((s: any) => ({
        id: s.server_id,
        name: s.server_name,
        ipv4: s.ipv4,
        status: s.status || 'UNKNOWN'
      }));

      setServers(formatted);
    } catch (error: any) {
      console.error("Failed to fetch servers", error);
      if (error.response && error.response.status === 400) {
        // Reset page if backend rejects pagination
        if (page > 1) {
          setPage(1);
        }
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchServers();
    const interval = setInterval(fetchServers, 5000);
    return () => clearInterval(interval);
  }, [page, pageSize]);

  const stats = {
    ONLINE: totalOnline,
    OFFLINE: totalOffline,
    UNKNOWN: totalUnknown,
  };
  const pieData = [
    { name: 'Online', value: stats.ONLINE, color: COLORS.ONLINE },
    { name: 'Offline', value: stats.OFFLINE, color: COLORS.OFFLINE },
    { name: 'Unknown', value: stats.UNKNOWN, color: COLORS.UNKNOWN },
  ];

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Live Dashboard</h1>
        <div style={{ color: 'var(--text-secondary)' }}>Total Servers: <strong>{total}</strong></div>
      </div>

      {isLoading ? <div style={{ textAlign: 'center', padding: '3rem' }}>Loading live data...</div> : (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: '2rem', marginBottom: '2rem' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div className="glass-panel" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div><h3 style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Online Servers</h3><div style={{ fontSize: '2rem', fontWeight: 700, color: COLORS.ONLINE }}>{stats.ONLINE}</div></div>
                <div className="status-indicator status-online pulse" style={{ width: '16px', height: '16px' }}></div>
              </div>
              <div className="glass-panel" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div><h3 style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Offline Servers</h3><div style={{ fontSize: '2rem', fontWeight: 700, color: COLORS.OFFLINE }}>{stats.OFFLINE}</div></div>
                <div className="status-indicator status-offline pulse" style={{ width: '16px', height: '16px' }}></div>
              </div>
              <div className="glass-panel" style={{ padding: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div><h3 style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Unknown Status</h3><div style={{ fontSize: '2rem', fontWeight: 700, color: COLORS.UNKNOWN }}>{stats.UNKNOWN}</div></div>
                <div className="status-indicator status-unknown" style={{ width: '16px', height: '16px' }}></div>
              </div>
            </div>

            <div className="glass-panel" style={{ padding: '1.5rem', minHeight: '300px' }}>
              <h3 style={{ marginBottom: '1rem', fontWeight: 600 }}>Status Distribution</h3>
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie data={pieData} cx="50%" cy="50%" innerRadius={60} outerRadius={80} paddingAngle={5} dataKey="value">
                    {pieData.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.color} />)}
                  </Pie>
                  <Tooltip contentStyle={{ backgroundColor: 'var(--bg-card)', borderColor: 'var(--border-light)' }} itemStyle={{ color: 'var(--text-primary)' }} />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '1.5rem' }}>Live Server Instances</h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(250px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
            {servers.map(server => (
              <div key={server.id} className={`glass-panel ${server.status === 'ONLINE' ? 'card-glow-online' : (server.status === 'OFFLINE' ? 'card-glow-offline' : '')}`} style={{ padding: '1.25rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <div style={{ fontWeight: 600, fontSize: '1rem' }}>{server.name}</div>
                  <div className={`status-indicator ${server.status === 'ONLINE' ? 'status-online' : (server.status === 'OFFLINE' ? 'status-offline' : 'status-unknown')} ${server.status !== 'UNKNOWN' ? 'pulse' : ''}`}></div>
                </div>
                <div style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>ID: {server.id}</div>
                <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', fontFamily: 'monospace' }}>{server.ipv4}</div>
              </div>
            ))}
          </div>

          {total > 0 && (
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem 1.5rem', background: 'var(--bg-card)', borderRadius: 'var(--radius-lg)', border: '1px solid var(--border-light)' }}>
              <span style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                Showing {Math.min((page - 1) * pageSize + 1, total)} to {Math.min(page * pageSize, total)} of {total} servers
              </span>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  className="btn btn-outline"
                  style={{ padding: '0.25rem 0.75rem' }}
                  disabled={page === 1}
                  onClick={() => setPage(p => p - 1)}>
                  Previous
                </button>
                <button
                  className="btn btn-outline"
                  style={{ padding: '0.25rem 0.75rem' }}
                  disabled={page * pageSize >= total}
                  onClick={() => setPage(p => p + 1)}>
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};
export default Dashboard;
