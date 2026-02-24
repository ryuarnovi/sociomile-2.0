import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Send, User, AlertTriangle, UserPlus, CheckCircle, Ticket } from 'lucide-react';
import { useAuthStore } from '../store/authStore';
import realtime from '../utils/realtime';
import { getConversation, listTicketsForConversation, setSelectedTicket, assignConversation, closeConversation } from '../api/conversationsService';
import { escalateTicket, listTickets } from '../api/ticketsService';
import { createMessage } from '../api/messagesService';

export function ConversationDetailPage() {
  const { id: _id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [message, setMessage] = useState('');
  const [messages, setMessages] = useState<any[]>([]);
  const [conversation, setConversation] = useState<any>(null);
  const [showEscalateModal, setShowEscalateModal] = useState(false);
  const [escalateForm, setEscalateForm] = useState({ priority: 'regular' });
  const [escalateTicketCode, setEscalateTicketCode] = useState('');
  const [allTickets, setAllTickets] = useState<any[]>([]);
  const [conversationTickets, setConversationTickets] = useState<any[]>([]);
  const [selectedTicketId, setSelectedTicketId] = useState<string | null>(null);
  const wsConnected = useRef(false);

  const handleSendMessage = async () => {
    if (!message.trim() || !conversation?.id) return;

    const text = message.trim();
    setMessage('');

    // optimistic UI
    const optimistic = {
      id: `msg_${Date.now()}`,
      sender_type: 'agent',
      sender_name: user?.name || 'Agent',
      message: text,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, optimistic]);

    try {
      const res = await createMessage(conversation.id, text);
      const created = res?.data || res;
      if (created) {
        setMessages((prev) => [...prev.filter((m) => m.id !== optimistic.id), created]);
      }
    } catch (err) {
      console.error('send failed', err);
    }
  };

  const handleAssignToMe = () => {
    if (!conversation?.id) return;
    (async () => {
      try {
        await assignConversation(conversation.id);
        setConversation({ ...conversation, status: 'assigned', assigned_agent_id: user?.id || '', assigned_agent_name: user?.name || '' });
        alert('Conversation assigned to you!');
      } catch (e) {
        console.error('assign failed', e);
        alert('Failed to assign conversation');
      }
    })();
  };

  const handleCloseConversation = () => {
    if (!conversation?.id) return;
    (async () => {
      try {
        await closeConversation(conversation.id);
        setConversation({ ...conversation, status: 'closed' });
        alert('Conversation closed!');
      } catch (e) {
        console.error('close failed', e);
        alert('Failed to close conversation');
      }
    })();
  };

  const handleEscalate = async () => {
    // allow escalate by existing ticket code OR by creating new ticket (title/description optional)
    // priority is required when no ticket code
    if (!escalateTicketCode && !escalateForm.priority) {
      alert('Please select a priority');
      return;
    }
    if (!conversation?.id) {
      alert('Conversation not loaded');
      return;
    }
    try {
      const payload: any = {};
      if (escalateTicketCode) {
        payload.ticket_code = escalateTicketCode;
      } else {
        // only send priority when creating a new ticket from escalate
        payload.priority = escalateForm.priority;
      }
      console.debug('escalate payload', payload);
      const res = await escalateTicket(conversation.id, payload);
      // if success, mark conversation as having a ticket
      if (res?.success) {
        const ticket = (res as any).data;
        setConversation({ ...conversation, has_ticket: true, ticket_id: ticket?.id });
        setShowEscalateModal(false);
        setEscalateTicketCode('');
        try {
          const t = await listTicketsForConversation(conversation.id);
          setConversationTickets(Array.isArray(t) ? t : (t.data || []));
        } catch (e) {}
        alert('Ticket processed successfully! (id: ' + (ticket?.id || '') + ')');
      } else {
        const msg = (res as any)?.message || 'Failed to create ticket';
        alert(msg);
      }
    } catch (err: any) {
      console.error('escalate failed', err, err?.response?.data);
      const serverMsg = err?.response?.data?.message || (err?.response?.data && JSON.stringify(err.response.data)) || null;
      const msg = serverMsg || err?.message || 'Failed to create ticket';
      alert(msg);
    }
  };

  // When opening escalate modal, fetch recent tickets for selection
  useEffect(() => {
    if (!showEscalateModal) return;
    let mounted = true;
    (async () => {
      try {
        const res = await listTickets({ page: 1, per_page: 50 });
        // res may be ApiResponse or array
        const items = res && Array.isArray((res as any).data) ? (res as any).data : (Array.isArray(res) ? res : []);
        if (!mounted) return;
        setAllTickets(items);
      } catch (e) {
        console.error('failed to load tickets for escalate select', e);
      }
    })();
    return () => { mounted = false; };
  }, [showEscalateModal]);

  useEffect(() => {
    if (!_id) return;
    let mounted = true;

    (async () => {
      try {
        const res = await getConversation(_id as string);
        const data = res ?? {};
        if (!mounted) return;
        const conv = data?.conversation ?? data;
        const normalizeTime = (v: any) => {
          if (!v) return '';
          try {
            let x = v;
            if (typeof x === 'object') {
              if (x.Time) x = x.Time;
              else if (x.String) x = x.String;
              else if (x.Valid === true && x.Time) x = x.Time;
              else x = String(x);
            }
            const d = new Date(x);
            if (isNaN(d.getTime())) return '';
            return d.toISOString();
          } catch (e) {
            return '';
          }
        };
        if (conv) conv.last_message_at = normalizeTime(conv.last_message_at);
        setConversation(conv);
        const msgs = Array.isArray(data?.messages) ? data.messages : [];
        setMessages(msgs);
        // load tickets for this conversation (may be empty)
        try {
          const t = await listTicketsForConversation(_id as string);
          setConversationTickets(Array.isArray(t) ? t : (t.data || []));
          if (conv?.ticket_id) setSelectedTicketId(conv.ticket_id);
        } catch (e) {
          // ignore
        }
      } catch (err) {
        console.error('failed to load conversation', err);
      }
    })();

    realtime.connect();
    wsConnected.current = true;

    const unsub = realtime.on('*', (msg) => {
      try {
        const payload = msg as any;
        const convId = String(payload.conversation_id || payload.conversationId || payload.conversation);
        if (convId === String(_id)) {
          // normalize incoming payload to expected message shape when possible
          setMessages((prev) => [...prev, payload]);
        }
      } catch (e) {
        // ignore
      }
    });

    return () => {
      mounted = false;
      try { unsub(); } catch (e) {}
      if (wsConnected.current) realtime.disconnect();
    };
  }, [_id]);

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {!conversation ? (
        <div className="p-6">Loading conversation...</div>
      ) : null}
      {/* Header */}
      <div className="bg-white rounded-t-xl shadow-sm p-4 flex items-center gap-4">
        <button onClick={() => navigate('/conversations')} className="p-2 hover:bg-gray-100 rounded-lg">
          <ArrowLeft className="w-5 h-5" />
        </button>
        <div className="w-10 h-10 bg-gray-200 rounded-full flex items-center justify-center">
          <User className="w-5 h-5 text-gray-500" />
        </div>
        <div className="flex-1">
          <h2 className="font-semibold text-gray-900">{conversation?.customer_name ?? 'Conversation'}</h2>
          <p className="text-sm text-gray-500">{conversation?.customer_external_id ?? ''} • {conversation?.channel ?? ''}</p>
        </div>
        <div className="flex items-center gap-2">
          <span className={`px-3 py-1 text-sm rounded-full ${
            (conversation?.status === 'open') ? 'bg-amber-100 text-amber-700' :
            (conversation?.status === 'assigned') ? 'bg-blue-100 text-blue-700' :
            'bg-gray-100 text-gray-700'
          }`}>
            {conversation?.status ?? 'unknown'}
          </span>
            <div className="flex items-center gap-4">
              {conversation?.has_ticket && (
                <span className="px-3 py-1 text-sm bg-purple-100 text-purple-700 rounded-full flex items-center gap-1">
                  <Ticket className="w-4 h-4" />
                  Escalated
                </span>
              )}
              <div>
                <label className="text-xs text-gray-500 mr-2">Assigned</label>
                <div className="text-sm text-gray-700">
                  {conversation?.assigned_agent_name || 'Unassigned'}
                </div>
              </div>
            </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 bg-gray-50 p-4 overflow-y-auto space-y-4">
        {messages.map((msg: any, idx) => {
          try {
            const senderType = msg.sender_type || msg.senderType || 'customer';
            const senderName = msg.sender_name || msg.senderName || (senderType === 'agent' ? (user?.name || 'Agent') : (conversation?.customer_name || 'Customer'));
            let createdAtRaw = msg.created_at || msg.createdAt || '';
            if (createdAtRaw && typeof createdAtRaw === 'object' && createdAtRaw.Time) {
              createdAtRaw = createdAtRaw.Time;
            }
            let createdAtStr = '';
            try {
              const d = new Date(createdAtRaw);
              if (!isNaN(d.getTime())) createdAtStr = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            } catch (e) {
              createdAtStr = '';
            }

            return (
              <div key={msg.id ?? `m_${idx}`} className={`flex ${senderType === 'agent' ? 'justify-end' : 'justify-start'}`}>
                <div className={`max-w-[70%] ${senderType === 'agent' ? 'bg-indigo-600 text-white rounded-l-xl rounded-tr-xl' : 'bg-white text-gray-900 rounded-r-xl rounded-tl-xl shadow-sm'} p-4`}>
                  <p className={`text-xs mb-1 ${senderType === 'agent' ? 'text-indigo-200' : 'text-gray-500'}`}>{senderName}</p>
                  <p>{String(msg.message ?? msg.text ?? '')}</p>
                  <p className={`text-xs mt-2 ${senderType === 'agent' ? 'text-indigo-200' : 'text-gray-400'}`}>{createdAtStr}</p>
                </div>
              </div>
            );
          } catch (e) {
            console.error('render message error', e, msg);
            return null;
          }
        })}
      </div>

      {/* Actions */}
      <div className="bg-white p-4 space-y-3">
        <div className="flex gap-2 flex-wrap">
          {conversation?.status === 'open' && (
            <button
              onClick={handleAssignToMe}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              <UserPlus className="w-4 h-4" />
              Assign to Me
            </button>
          )}
          {conversation && conversation.status !== 'closed' && (
            <>
              <button
                onClick={handleCloseConversation}
                className="flex items-center gap-2 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700"
              >
                <CheckCircle className="w-4 h-4" />
                Close
              </button>
              {conversation && !conversation.has_ticket && (
                <button
                  onClick={() => setShowEscalateModal(true)}
                  className="flex items-center gap-2 px-4 py-2 bg-amber-600 text-white rounded-lg hover:bg-amber-700"
                >
                  <AlertTriangle className="w-4 h-4" />
                  Escalate to Ticket
                </button>
              )}
            </>
          )}
        </div>

        {/* Message Input */}
        {conversation && conversation.status !== 'closed' && (
          <div className="flex gap-2">
            <input
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
              placeholder="Type your message..."
              className="flex-1 px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
            />
            <button
              onClick={handleSendMessage}
              className="px-4 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
            >
              <Send className="w-5 h-5" />
            </button>
          </div>
        )}
      </div>

      {/* Escalate Modal */}
      {showEscalateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
            <h3 className="text-lg font-semibold mb-4">Escalate to Ticket</h3>
            <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Use existing ticket (optional)</label>
                  <select
                    value={escalateTicketCode}
                    onChange={(e) => {
                      const val = e.target.value || '';
                      setEscalateTicketCode(val);
                    }}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                  >
                    <option value="">-- select existing ticket (optional) --</option>
                    {allTickets.map((t) => (
                      <option key={t.id} value={t.code || t.id}>{t.title}{t.code ? ` (${t.code})` : ` (${t.id})`}</option>
                    ))}
                  </select>
                </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Priority</label>
                <select
                  value={escalateForm.priority}
                  onChange={(e) => setEscalateForm({ ...escalateForm, priority: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                >
                  <option value="stabdar">Standar</option>
                  <option value="regular">Regular</option>
                  <option value="VIP">VIP</option>
                </select>
              </div>
            </div>
            <div className="flex gap-2 mt-6">
              <button
                onClick={() => setShowEscalateModal(false)}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleEscalate}
                className="flex-1 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700"
              >
                Create Ticket
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
