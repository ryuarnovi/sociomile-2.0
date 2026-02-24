import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Filter, User } from 'lucide-react';
import { listConversations } from '../api/conversationsService';

export function ConversationsPage() {
  const navigate = useNavigate();
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [conversations, setConversations] = useState<any[]>([]);

  useEffect(() => {
    listConversations().then((data) => {
      if (Array.isArray(data)) setConversations(data);
    }).catch(() => {});
  }, []);

  const filteredConversations = conversations.filter(conv => {
    if (statusFilter !== 'all' && conv.status !== statusFilter) return false;
    if (searchQuery && !(conv.customer_name || '').toLowerCase().includes(searchQuery.toLowerCase())) return false;
    return true;
  });

  const getStatusBadge = (status: string) => {
    const styles = {
      open: 'bg-amber-100 text-amber-700',
      assigned: 'bg-blue-100 text-blue-700',
      closed: 'bg-gray-100 text-gray-700',
    };
    return styles[status as keyof typeof styles] || styles.open;
  };

  const getChannelColor = (channel: string) => {
    const colors: Record<string, string> = {
      WhatsApp: 'bg-green-500',
      Instagram: 'bg-pink-500',
      Telegram: 'bg-blue-500',
      Email: 'bg-gray-500',
    };
    return colors[channel] || 'bg-gray-500';
  };

  const formatTimeSafe = (value: any) => {
    if (!value) return '';
    try {
      let v = value;
      // unwrap common Go sql.NullTime / custom shapes
      if (typeof v === 'object') {
        if (v.Time) v = v.Time;
        else if (v.String) v = v.String;
        else if (v.Valid === true && v.Time) v = v.Time;
        else v = String(v);
      }
      const d = new Date(v);
      if (isNaN(d.getTime())) return '';
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch (e) {
      return '';
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Conversations</h1>
        <p className="text-gray-500">Manage customer conversations across all channels</p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search by customer name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-5 h-5 text-gray-400" />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
          >
            <option value="all">All Status</option>
            <option value="open">Open</option>
            <option value="assigned">Assigned</option>
            <option value="closed">Closed</option>
          </select>
        </div>
      </div>

      {/* Conversation List */}
      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <div className="divide-y divide-gray-100">
          {filteredConversations.map((conv) => (
            <div
              key={conv.id}
              onClick={() => navigate(`/conversations/${conv.id}`)}
              className="p-4 hover:bg-gray-50 cursor-pointer transition"
            >
              <div className="flex items-start gap-4">
                <div className="relative">
                  <div className="w-12 h-12 bg-gray-200 rounded-full flex items-center justify-center">
                    <User className="w-6 h-6 text-gray-500" />
                  </div>
                  <div className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-white ${getChannelColor(conv.channel)}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-gray-900">{conv.customer_name}</h3>
                    <span className={`px-2 py-0.5 text-xs rounded-full ${getStatusBadge(conv.status)}`}>
                      {conv.status}
                    </span>
                  </div>
                  <p className="text-sm text-gray-500">{conv.customer_external_id} • {conv.channel}</p>
                  <p className="text-sm text-gray-700 mt-1 truncate">{conv.last_message}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm text-gray-400">{formatTimeSafe(conv.last_message_at)}</p>
                  {conv.assigned_agent_name && (
                    <p className="text-xs text-gray-500 mt-1">
                      Assigned: {conv.assigned_agent_name}
                    </p>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
