import React, { useState, useEffect } from 'react';
import { Plus, Download, Upload, FileText, Trash2, Edit } from 'lucide-react';
import api from '../api/axios';

interface Server { id: string; name: string; ipv4: string; }

const ServerManagement: React.FC = () => {
  const [servers, setServers] = useState<Server[]>([]);
  const [isReportModalOpen, setReportModalOpen] = useState(false);
  const [reportEmail, setReportEmail] = useState('');
  const [reportFrom, setReportFrom] = useState('');
  const [reportTo, setReportTo] = useState('');
  const [importResult, setImportResult] = useState<any>(null);
  const [toastMessage, setToastMessage] = useState<{ text: string, type: 'error' | 'success' } | null>(null);

  const showToast = (text: string, type: 'error' | 'success' = 'error') => {
    setToastMessage({ text, type });
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleError = (defaultMsg: string, err: any) => {
    const apiMsg = err.response?.data?.message || err.response?.data?.msg;
    showToast(apiMsg ? `${defaultMsg}: ${apiMsg}` : defaultMsg, 'error');
  };

  const [isServerModalOpen, setServerModalOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<Server | null>(null);
  const [serverFormData, setServerFormData] = useState({ id: '', name: '', ipv4: '' });
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  const fetchServers = async () => {
    try {
      const from = (page - 1) * pageSize;
      const to = page * pageSize;
      const res = await api.get('/servers', {
        params: { from, to, sort_field: 'server_name', desc: false }
      });
      const items = res.data.items || [];
      setTotal(res.data.total || 0);
      const formatted = items.map((s: any) => ({
        id: s.server_id,
        name: s.server_name,
        ipv4: s.ipv4
      }));
      setServers(formatted);
    } catch (e) {
      console.error("Failed to fetch servers", e);
    }
  };

  useEffect(() => { fetchServers(); }, [page]);

  const confirmDelete = async () => {
    if (!deleteConfirmId) return;
    try {
      await api.delete(`/servers/${deleteConfirmId}`);
      setDeleteConfirmId(null);
      fetchServers();
      showToast('Server deleted successfully', 'success');
    } catch (e) { handleError('Failed to delete server', e); }
  };

  const openAddModal = () => {
    setEditingServer(null);
    setServerFormData({ id: '', name: '', ipv4: '' });
    setServerModalOpen(true);
  };

  const openEditModal = (server: Server) => {
    setEditingServer(server);
    setServerFormData({ id: server.id, name: server.name, ipv4: server.ipv4 });
    setServerModalOpen(true);
  };

  const handleSaveServer = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingServer) {
        await api.patch(`/servers/${editingServer.id}`, { server_name: serverFormData.name, ipv4: serverFormData.ipv4 });
      } else {
        await api.post('/servers', { server_id: serverFormData.id, server_name: serverFormData.name, ipv4: serverFormData.ipv4 });
      }
      setServerModalOpen(false);
      fetchServers();
      showToast(editingServer ? 'Server updated successfully' : 'Server created successfully', 'success');
    } catch (err) {
      handleError('Failed to save server', err);
    }
  };

  const handleRequestReport = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await api.post('/monitor/report', { 
        from: new Date(reportFrom).toISOString(), 
        to: new Date(reportTo).toISOString(), 
        receivers: [reportEmail] 
      });
      showToast('Report request accepted!', 'success');
      setReportModalOpen(false);
    } catch (e) { handleError('Failed to request report', e); }
  };

  const handleImport = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const formData = new FormData();
    formData.append('file', file);
    try {
      const res = await api.post('/servers/import', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
      setImportResult(res.data);
      fetchServers();
    } catch (err) { handleError('Import failed', err); }
  };

  const handleExport = async () => {
    try {
      const res = await api.get('/servers/export', {
        params: { from: 0, to: Math.max(total, 1), sort_field: 'server_name', desc: false },
        responseType: 'blob'
      });
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'servers.xlsx');
      document.body.appendChild(link);
      link.click();
      link.remove();
      showToast('Export successful', 'success');
    } catch (err) { handleError('Export failed', err); }
  };

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">Server Inventory</h1>
        <div style={{ display: 'flex', gap: '0.75rem' }}>
          <label className="btn btn-outline" style={{ cursor: 'pointer' }}><Upload size={16} />Import CSV<input type="file" accept=".csv,.xlsx" hidden onChange={handleImport} /></label>
          <button className="btn btn-outline" onClick={handleExport}><Download size={16} />Export</button>
          <button className="btn btn-primary" onClick={() => setReportModalOpen(true)}><FileText size={16} />Request Report</button>
          <button className="btn btn-primary" style={{ background: 'var(--accent-success)' }} onClick={openAddModal}><Plus size={16} />Add Server</button>
        </div>
      </div>
      <div className="glass-panel" style={{ overflow: 'hidden' }}>
        <table className="data-table">
          <thead><tr><th>Server ID</th><th>Name</th><th>IP Address</th><th style={{ textAlign: 'right' }}>Actions</th></tr></thead>
          <tbody>
            {servers.length === 0 ? <tr><td colSpan={4} style={{ textAlign: 'center', padding: '3rem', color: 'var(--text-muted)' }}>No servers found.</td></tr> : servers.map(server => (
              <tr key={server.id}>
                <td style={{ fontFamily: 'monospace', color: 'var(--accent-primary)' }}>{server.id}</td>
                <td style={{ fontWeight: 500 }}>{server.name}</td><td style={{ color: 'var(--text-secondary)' }}>{server.ipv4}</td>
                <td style={{ textAlign: 'right' }}>
                  <button className="btn btn-outline" style={{ padding: '0.5rem', marginRight: '0.5rem' }} onClick={() => openEditModal(server)}><Edit size={16} /></button>
                  <button className="btn btn-danger" style={{ padding: '0.5rem' }} onClick={() => setDeleteConfirmId(server.id)}><Trash2 size={16} /></button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {total > 0 && (
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1rem 1.5rem', borderTop: '1px solid var(--border-light)' }}>
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
      </div>
      {isReportModalOpen && (
        <div className="modal-overlay" onClick={() => setReportModalOpen(false)}>
          <div className="glass-panel modal-content" onClick={e => e.stopPropagation()} style={{ padding: '2rem' }}>
            <div className="modal-header"><h2 className="modal-title">Request Status Report</h2><button className="modal-close" onClick={() => setReportModalOpen(false)}>✕</button></div>
            <form onSubmit={handleRequestReport}>
              <div className="input-group"><label className="input-label">Email to receive report</label><input type="email" className="input-field" required value={reportEmail} onChange={e => setReportEmail(e.target.value)} /></div>
              <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem' }}>
                <div style={{ flex: 1 }}><label className="input-label">From Date</label><input type="date" className="input-field" required value={reportFrom} onChange={e => setReportFrom(e.target.value)} /></div>
                <div style={{ flex: 1 }}><label className="input-label">To Date</label><input type="date" className="input-field" required value={reportTo} onChange={e => setReportTo(e.target.value)} /></div>
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem' }}><button type="button" className="btn btn-outline" onClick={() => setReportModalOpen(false)}>Cancel</button><button type="submit" className="btn btn-primary">Submit</button></div>
            </form>
          </div>
        </div>
      )}
      {importResult && (
        <div className="modal-overlay" onClick={() => setImportResult(null)}>
          <div className="glass-panel modal-content" onClick={e => e.stopPropagation()} style={{ padding: '2rem' }}>
            <div className="modal-header"><h2 className="modal-title">Import Results</h2><button className="modal-close" onClick={() => setImportResult(null)}>✕</button></div>
            <div style={{ marginTop: '1rem', marginBottom: '1.5rem' }}>
              <p style={{ color: 'var(--accent-success)', fontWeight: 'bold', fontSize: '1.1rem' }}>Success: {importResult.total_success}</p>
              {importResult.id_success && importResult.id_success.length > 0 && (
                <div style={{ maxHeight: '120px', overflowY: 'auto', background: 'var(--bg-tertiary)', padding: '0.75rem', borderRadius: '8px', fontSize: '0.85rem', marginTop: '0.5rem', marginBottom: '1.25rem', border: '1px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                  {importResult.id_success.join(', ')}
                </div>
              )}
              
              <p style={{ color: 'var(--accent-danger)', fontWeight: 'bold', fontSize: '1.1rem' }}>Failed: {importResult.total_failed}</p>
              {importResult.id_failed && importResult.id_failed.length > 0 && (
                <div style={{ maxHeight: '120px', overflowY: 'auto', background: 'var(--bg-tertiary)', padding: '0.75rem', borderRadius: '8px', fontSize: '0.85rem', marginTop: '0.5rem', border: '1px solid var(--border-color)', color: 'var(--text-secondary)' }}>
                  {importResult.id_failed.join(', ')}
                </div>
              )}
            </div>
            <div style={{ display: 'flex', justifyContent: 'flex-end' }}><button className="btn btn-primary" onClick={() => setImportResult(null)}>Close</button></div>
          </div>
        </div>
      )}

      {isServerModalOpen && (
        <div className="modal-overlay" onClick={() => setServerModalOpen(false)}>
          <div className="glass-panel modal-content" onClick={e => e.stopPropagation()} style={{ padding: '2rem' }}>
            <div className="modal-header">
              <h2 className="modal-title">{editingServer ? 'Edit Server' : 'Add Server'}</h2>
              <button className="modal-close" onClick={() => setServerModalOpen(false)}>✕</button>
            </div>
            <form onSubmit={handleSaveServer}>
              <div className="input-group">
                <label className="input-label">Server ID</label>
                <input type="text" className="input-field" required disabled={!!editingServer} value={serverFormData.id} onChange={e => setServerFormData({ ...serverFormData, id: e.target.value })} placeholder="server_001" />
              </div>
              <div className="input-group">
                <label className="input-label">Server Name</label>
                <input type="text" className="input-field" required value={serverFormData.name} onChange={e => setServerFormData({ ...serverFormData, name: e.target.value })} placeholder="Database Server" />
              </div>
              <div className="input-group">
                <label className="input-label">IPv4 Address</label>
                <input type="text" className="input-field" required pattern="^([0-9]{1,3}\.){3}[0-9]{1,3}$" value={serverFormData.ipv4} onChange={e => setServerFormData({ ...serverFormData, ipv4: e.target.value })} placeholder="192.168.1.100" />
              </div>
              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem', marginTop: '1.5rem' }}>
                <button type="button" className="btn btn-outline" onClick={() => setServerModalOpen(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary">Save</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {deleteConfirmId && (
        <div className="modal-overlay" onClick={() => setDeleteConfirmId(null)}>
          <div className="glass-panel modal-content" onClick={e => e.stopPropagation()} style={{ padding: '2rem', maxWidth: '400px' }}>
            <h2 className="modal-title" style={{ color: 'var(--accent-danger)', marginBottom: '1rem' }}>Confirm Delete</h2>
            <p style={{ color: 'var(--text-secondary)', marginBottom: '2rem' }}>Are you sure you want to delete server <strong>{deleteConfirmId}</strong>? This action cannot be undone.</p>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '1rem' }}>
              <button type="button" className="btn btn-outline" onClick={() => setDeleteConfirmId(null)}>Cancel</button>
              <button type="button" className="btn btn-danger" onClick={confirmDelete}>Delete</button>
            </div>
          </div>
        </div>
      )}

      {toastMessage && (
        <div style={{
          position: 'fixed',
          top: '20px',
          right: '20px',
          background: toastMessage.type === 'error' ? 'var(--accent-danger)' : 'var(--accent-success)',
          color: 'white',
          padding: '1rem 1.5rem',
          borderRadius: '8px',
          boxShadow: '0 4px 12px rgba(0,0,0,0.2)',
          zIndex: 9999,
          fontWeight: 500,
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
          animation: 'fadeIn 0.3s ease-out',
        }}>
          {toastMessage.text}
        </div>
      )}
    </div>
  );
};
export default ServerManagement;
