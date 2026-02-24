/* eslint-disable @typescript-eslint/no-unused-vars */
/* eslint-disable no-restricted-globals */
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Filter, AlertTriangle } from 'lucide-react';
import { AxiosError } from 'axios';
import { useAuthStore } from '../store/authStore';
import { listTickets, escalateTicket, updateTicketStatus, deleteTicket, createTicket, updateTicket } from '../api/ticketsService';
import { listConversations, setSelectedTicket as setConversationSelectedTicket } from '../api/conversationsService';
import { Ticket } from '../types';

export function TicketsPage() {
  const { user } = useAuthStore();
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [priorityFilter, setPriorityFilter] = useState<string>('all');
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null);
  const [newStatus, setNewStatus] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showEdit, setShowEdit] = useState(false);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [newConversationId, setNewConversationId] = useState('');
  const [conversations, setConversations] = useState<any[]>([]);
  const [newPriority, setNewPriority] = useState('low');
  const navigate = useNavigate();

  const fetchTickets = async () => {
    setLoading(true);
    try {
      const params: any = {};
      if (statusFilter !== 'all') params.status = statusFilter;
      if (priorityFilter !== 'all') params.priority = priorityFilter;
      const res = await listTickets(params);
      // Normalize response: ApiResponse { success, data: [...] } or direct array
      if (res && Array.isArray((res as any).data)) setTickets((res as any).data);
      else if (Array.isArray(res)) setTickets(res as any);
      else setTickets([]);
    } catch (err) {
      console.error('Failed to fetch tickets', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTickets();
    // fetch small list of conversations for create modal
    (async () => {
      try {
        const convs = await listConversations();
        // normalize response to an array
        if (Array.isArray(convs)) setConversations(convs);
        else if (convs && Array.isArray((convs as any).data)) setConversations((convs as any).data);
        else setConversations([]);
      } catch (e) {
        console.error('failed to fetch conversations', e);
      }
    })();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter, priorityFilter]);

  const getStatusBadge = (status: string) => {
    const styles: Record<string, string> = {
      open: 'bg-red-100 text-red-700',
      in_progress: 'bg-amber-100 text-amber-700',
      resolved: 'bg-emerald-100 text-emerald-700',
      closed: 'bg-gray-100 text-gray-700',
    };
    return styles[status] || styles.open;
  };

  const getPriorityBadge = (priority: string) => {
    const styles: Record<string, string> = {
      stabdar: 'bg-gray-100 text-gray-700',
      regular: 'bg-blue-100 text-blue-700',
      VIP: 'bg-red-100 text-red-700',
    };
    return styles[priority] || styles.stabdar;
  };

  const handleUpdateStatus = async () => {
    if (!selectedTicket || !newStatus) return;
    try {
      await updateTicketStatus(selectedTicket.id, newStatus);
      setSelectedTicket(null);
      setNewStatus('');
      fetchTickets();
      alert('Ticket status updated!');
    } catch (err) {
      console.error(err);
      alert('Failed to update status');
    }
  };

  const handleCreate = async () => {
    if (!newTitle) {
      alert('Title is required');
      return;
    }
    try {
      // `newConversationId` now holds optional ticket code; create ticket with code
      const payload: any = { title: newTitle, description: newDescription, priority: newPriority };
      if (newConversationId) payload.code = newConversationId;
      // leave conversation_id empty for tickets created here
      await createTicket(payload);
      setShowCreate(false);
      setNewTitle('');
      setNewDescription('');
      setNewConversationId('');
      setNewPriority('low');
      fetchTickets();
      alert('Ticket created');
    } catch (err) {
      console.error(err);
      const axiosError = err as AxiosError<{ message: string }>;
      const msg = axiosError?.response?.data?.message || axiosError?.message || 'Failed to create ticket';
      alert(msg);
    }
  };

  const handleEditSave = async () => {
    if (!selectedTicket) return;
    try {
      // update title/description/priority
      // If user set a code, update ticket.code; otherwise no change to conversation mapping here
      const payload: any = { title: newTitle, description: newDescription, priority: newPriority };
      if (newConversationId) payload.code = newConversationId;
      await updateTicket(selectedTicket.id, payload);

      setShowEdit(false);
      setSelectedTicket(null);
      setNewTitle('');
      setNewDescription('');
      setNewConversationId('');
      setNewPriority('low');
      fetchTickets();
      alert('Ticket updated');
    } catch (err) {
      console.error(err);
      alert('Failed to update ticket');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this ticket?')) return;
    try {
      await deleteTicket(id);
      fetchTickets();
      alert('Ticket deleted');
    } catch (err) {
      console.error(err);
      alert('Failed to delete ticket');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Tickets</h1>
          <p className="text-gray-500">Manage escalated issues and support tickets</p>
        </div>
        <div>
          <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-indigo-600 text-white rounded-lg">Create Ticket</button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex items-center gap-2">
          <Filter className="w-5 h-5 text-gray-400" />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
          >
            <option value="all">All Status</option>
            <option value="open">Open</option>
            <option value="in_progress">In Progress</option>
            <option value="resolved">Resolved</option>
            <option value="closed">Closed</option>
          </select>
        </div>
        <div className="flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-gray-400" />
          <select
            value={priorityFilter}
            onChange={(e) => setPriorityFilter(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
          >
            <option value="all">All Priority</option>
            <option value="stabdar">Standar</option>
            <option value="regular">Regular</option>
            <option value="VIP">VIP</option>
          </select>
        </div>
      </div>

      {/* Tickets Table */}
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ticket</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Priority</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Assigned To</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading && (
                <tr>
                  <td colSpan={6} className="px-6 py-4 text-center text-sm text-gray-500">Loading...</td>
                </tr>
              )}
              {!loading && tickets.map((ticket) => (
                <tr key={ticket.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <div>
                      <p className="font-medium text-gray-900">{ticket.title}</p>
                      <p className="text-sm text-gray-500">{ticket.description}</p>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadge(ticket.status)}`}>
                      {ticket.status.replace('_', ' ')}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 text-xs rounded-full ${getPriorityBadge(ticket.priority)}`}>
                      {ticket.priority}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">{ticket.assigned_agent_name}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">
                    {new Date(ticket.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 space-x-2">
                    {user?.role === 'admin' && ticket.status !== 'closed' && (
                      <button
                        onClick={() => {
                                          setSelectedTicket(ticket);
                                          setNewStatus(ticket.status);
                                          setShowStatusModal(true);
                        }}
                        className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
                      >
                        Update Status
                      </button>
                    )}
                    {user?.role === 'admin' && (
                      <button
                        onClick={() => {
                          setSelectedTicket(ticket);
                          setNewTitle(ticket.title || '');
                          setNewDescription(ticket.description || '');
                          setNewPriority(ticket.priority || 'low');
                          setNewConversationId((ticket as any).code || '');
                          setShowEdit(true);
                        }}
                        className="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
                      >
                        Edit
                      </button>
                    )}
                    {user?.role === 'admin' && (
                      <button onClick={() => handleDelete(ticket.id)} className="text-red-600 hover:text-red-800 text-sm font-medium">Delete</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Ticket Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
            <h3 className="text-lg font-semibold mb-4">Create Ticket</h3>
            <div className="space-y-3 mb-4">
              <input value={newTitle} onChange={(e) => setNewTitle(e.target.value)} placeholder="Title" className="w-full px-4 py-2 border rounded" />
              <textarea value={newDescription} onChange={(e) => setNewDescription(e.target.value)} placeholder="Description" className="w-full px-4 py-2 border rounded" />
              <select value={newPriority} onChange={(e) => setNewPriority(e.target.value)} className="w-full px-4 py-2 border rounded">
                <option value="stabdar">Standar</option>
                <option value="regular">Regular</option>
                <option value="VIP">VIP</option>
              </select>
              <input value={newConversationId} onChange={(e) => setNewConversationId(e.target.value)} placeholder="Ticket code (optional)" className="w-full px-4 py-2 border rounded" />
            </div>
            <div className="flex gap-2">
              <button onClick={() => setShowCreate(false)} className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
              <button onClick={handleCreate} className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg">Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Ticket Modal */}
      {showEdit && selectedTicket && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
            <h3 className="text-lg font-semibold mb-4">Edit Ticket</h3>
            <div className="space-y-3 mb-4">
              <input value={newTitle} onChange={(e) => setNewTitle(e.target.value)} placeholder="Title" className="w-full px-4 py-2 border rounded" />
              <textarea value={newDescription} onChange={(e) => setNewDescription(e.target.value)} placeholder="Description" className="w-full px-4 py-2 border rounded" />
              <select value={newPriority} onChange={(e) => setNewPriority(e.target.value)} className="w-full px-4 py-2 border rounded">
                <option value="stabdar">Standar</option>
                <option value="regular">Regular</option>
                <option value="VIP">VIP</option>
              </select>
              <input value={newConversationId} onChange={(e) => setNewConversationId(e.target.value)} placeholder="Ticket code (optional)" className="w-full px-4 py-2 border rounded" />
            </div>
            <div className="flex gap-2">
              <button onClick={() => { setShowEdit(false); setSelectedTicket(null); }} className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
              <button onClick={handleEditSave} className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg">Save</button>
            </div>
          </div>
        </div>
      )}

      {/* Update Status Modal */}
      {selectedTicket && showStatusModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
            <h3 className="text-lg font-semibold mb-4">Update Ticket Status</h3>
            <p className="text-gray-600 mb-4">{selectedTicket.title}</p>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
              <select
                value={newStatus}
                onChange={(e) => setNewStatus(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
              >
                <option value="open">Open</option>
                <option value="in_progress">In Progress</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => { setSelectedTicket(null); setShowStatusModal(false); }}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={async () => { await handleUpdateStatus(); setShowStatusModal(false); }}
                className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
              >
                Update
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
