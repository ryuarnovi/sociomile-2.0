import { useEffect, useState } from 'react';
import { Radio, Send, MessageSquare } from 'lucide-react';
import { useAuthStore } from '../store/authStore';
import client from '../api/client';
import { listChannels, createChannel } from '../api/channelsService';

export function ChannelSimulatorPage() {
  const { user } = useAuthStore();
  const [formData, setFormData] = useState({
    tenant_id: user?.tenant_id || 'tenant_001',
    customer_external_id: '',
    channel: 'WhatsApp',
    message: '',
  });
  const [responses, setResponses] = useState<Array<{ success: boolean; message: string; timestamp: string }>>([]);
  const [loading, setLoading] = useState(false);
  const [channels, setChannels] = useState<any[]>([]);
  const [selectedChannel, setSelectedChannel] = useState<string>('WhatsApp');
  const [newChannelName, setNewChannelName] = useState('');

  // keep formData.channel in sync with selectedChannel for preview and payload
  useEffect(() => {
    setFormData((f) => ({ ...f, channel: selectedChannel }));
  }, [selectedChannel]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const payload = {
        tenant_id: formData.tenant_id,
        customer_external_id: formData.customer_external_id,
        channel: selectedChannel,
        message: formData.message,
      };

      const res = await client.post('/channel/webhook', payload);
      const ok = res?.data?.success ?? true;
      const msg = res?.data?.message ?? 'Webhook delivered';
      const response = { success: ok, message: msg, timestamp: new Date().toISOString() };
      setResponses([response, ...responses]);
      setFormData({ ...formData, message: '' });
    } catch (err: any) {
      const response = { success: false, message: err?.response?.data?.message ?? err.message, timestamp: new Date().toISOString() };
      setResponses([response, ...responses]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    listChannels().then((data) => {
      if (Array.isArray(data)) setChannels(data);
    }).catch(() => {});
  }, []);

  // when channels are loaded, default selectedChannel to first channel (prefer name then slug)
  useEffect(() => {
    if (channels.length > 0) {
      const first = channels[0];
      const val = first.name || first.slug || first.id;
      setSelectedChannel(val);
    }
  }, [channels]);

  const handleCreateChannel = async () => {
    if (!newChannelName.trim()) return;
    try {
      const slug = newChannelName.trim().toLowerCase().replace(/\s+/g, '-');
      const created = await createChannel({ name: newChannelName.trim(), slug });
      setChannels((prev) => [created, ...prev]);
      setSelectedChannel(created.slug || created.name);
      // notify other pages to refresh channel lists
      try { window.dispatchEvent(new CustomEvent('channels:changed')); } catch (e) {}
      setNewChannelName('');
    } catch (err) {
      // ignore for now
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Channel Simulator</h1>
        <p className="text-gray-500">Simulate incoming messages from external channels</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Simulator Form */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-3 bg-indigo-100 rounded-lg">
              <Radio className="w-6 h-6 text-indigo-600" />
            </div>
            <div>
              <h2 className="font-semibold text-gray-900">Webhook Simulator</h2>
              <p className="text-sm text-gray-500">POST /channel/webhook</p>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tenant ID</label>
              <input
                type="text"
                value={formData.tenant_id}
                onChange={(e) => setFormData({ ...formData, tenant_id: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Customer External ID</label>
              <input
                type="text"
                value={formData.customer_external_id}
                onChange={(e) => setFormData({ ...formData, customer_external_id: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                placeholder="e.g., +6281234567890 or @username"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Channel</label>
              <div className="flex flex-col sm:flex-row gap-2">
                <select
                  value={selectedChannel}
                  onChange={(e) => {
                    const v = e.target.value;
                    setSelectedChannel(v);
                    setFormData((f) => ({ ...f, channel: v }));
                  }}
                  className="w-full sm:flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                >
                  {channels.length > 0 ? (
                    channels.map((ch) => <option key={ch.id} value={ch.name || ch.slug || ch.id}>{ch.name || ch.slug || ch.id}</option>)
                  ) : (
                    <>
                      <option value="WhatsApp">WhatsApp</option>
                      <option value="Instagram">Instagram</option>
                      <option value="Telegram">Telegram</option>
                      <option value="Email">Email</option>
                      <option value="Twitter">Twitter</option>
                    </>
                  )}
                </select>
                <input
                  type="text"
                  placeholder="New channel"
                  value={newChannelName}
                  onChange={(e) => setNewChannelName(e.target.value)}
                  className="w-full sm:w-auto px-3 py-2 border border-gray-300 rounded-lg"
                />
                <button type="button" onClick={handleCreateChannel} className="w-full sm:w-auto px-3 py-2 bg-indigo-600 text-white rounded-lg">Create</button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Message</label>
              <textarea
                value={formData.message}
                onChange={(e) => setFormData({ ...formData, message: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
                rows={4}
                placeholder="Enter customer message..."
                required
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full flex items-center justify-center gap-2 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50"
            >
              <Send className="w-5 h-5" />
              {loading ? 'Sending...' : 'Send Webhook'}
            </button>
          </form>

          {/* Request Preview */}
          <div className="mt-6 p-4 bg-gray-900 rounded-lg">
            <p className="text-xs text-gray-400 mb-2">Request Payload Preview:</p>
            <pre className="text-sm text-green-400 overflow-auto">
{JSON.stringify({
  tenant_id: formData.tenant_id,
  customer_external_id: formData.customer_external_id,
  channel: selectedChannel,
  message: formData.message || "..."
}, null, 2)}
            </pre>
          </div>
        </div>

        {/* Response Log */}
        <div className="bg-white rounded-xl shadow-sm p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="p-3 bg-emerald-100 rounded-lg">
              <MessageSquare className="w-6 h-6 text-emerald-600" />
            </div>
            <div>
              <h2 className="font-semibold text-gray-900">Response Log</h2>
              <p className="text-sm text-gray-500">Recent webhook responses</p>
            </div>
          </div>

          <div className="space-y-3 max-h-[500px] overflow-y-auto">
            {responses.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Radio className="w-12 h-12 mx-auto mb-3 opacity-50" />
                <p>No responses yet</p>
                <p className="text-sm">Send a webhook to see responses here</p>
              </div>
            ) : (
              responses.map((response, index) => (
                <div
                  key={index}
                  className={`p-4 rounded-lg ${
                    response.success ? 'bg-emerald-50 border border-emerald-200' : 'bg-red-50 border border-red-200'
                  }`}
                >
                  <div className="flex items-center gap-2 mb-2">
                    <span className={`w-2 h-2 rounded-full ${response.success ? 'bg-emerald-500' : 'bg-red-500'}`} />
                    <span className={`text-sm font-medium ${response.success ? 'text-emerald-700' : 'text-red-700'}`}>
                      {response.success ? 'Success' : 'Error'}
                    </span>
                    <span className="text-xs text-gray-500 ml-auto">
                      {new Date(response.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <p className="text-sm text-gray-700">{response.message}</p>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
