import { MessageSquare, Ticket, Users, Clock, CheckCircle, AlertCircle, TrendingUp } from 'lucide-react';
import { useEffect, useState } from 'react';
import { listConversations } from '../api/conversationsService';
import { listTickets } from '../api/ticketsService';

type ConvCounts = { total: number; open: number; assigned: number; closed: number };
type TicketCounts = { total: number; open: number; in_progress: number; resolved: number; closed: number };

export function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [convCounts, setConvCounts] = useState<ConvCounts>({ total: 0, open: 0, assigned: 0, closed: 0 });
  const [ticketCounts, setTicketCounts] = useState<TicketCounts>({ total: 0, open: 0, in_progress: 0, resolved: 0, closed: 0 });

  useEffect(() => {
    let mounted = true;
    async function fetchStats() {
      setLoading(true);
      try {
        const [convsRes, ticketsRes] = await Promise.all([
          listConversations(),
          listTickets({ per_page: 100 })
        ]);

        if (!mounted) return;

        const convs = Array.isArray(convsRes) ? convsRes : (convsRes?.data ?? []);
        const convTotal = convs.length;
        const convOpen = convs.filter((c: any) => c.status === 'open').length;
        const convAssigned = convs.filter((c: any) => c.status === 'assigned' || (c.assigned_agent_id && c.assigned_agent_id !== '')).length;
        const convClosed = convs.filter((c: any) => c.status === 'closed').length;

        setConvCounts({ total: convTotal, open: convOpen, assigned: convAssigned, closed: convClosed });

        const ticketsPayload = ticketsRes?.data ?? ticketsRes?.data?.data ?? ticketsRes?.data ?? ticketsRes;
        const tickets = Array.isArray(ticketsPayload) ? ticketsPayload : (ticketsRes?.data?.data ?? []);
        // if API returned ApiResponse with data
        const tlist = Array.isArray(tickets) ? tickets : (ticketsRes?.data?.data ?? []);
        const tTotal = tlist.length;
        const tOpen = tlist.filter((t: any) => t.status === 'open').length;
        const tInProg = tlist.filter((t: any) => t.status === 'in_progress').length;
        const tResolved = tlist.filter((t: any) => t.status === 'resolved').length;
        const tClosed = tlist.filter((t: any) => t.status === 'closed').length;

        setTicketCounts({ total: tTotal, open: tOpen, in_progress: tInProg, resolved: tResolved, closed: tClosed });
      } catch (e) {
        // keep defaults
      } finally {
        if (mounted) setLoading(false);
      }
    }
    fetchStats();
    return () => { mounted = false; };
  }, []);

  const statColor = (color: string) => ({ indigo: 'bg-indigo-100 text-indigo-600', amber: 'bg-amber-100 text-amber-600', blue: 'bg-blue-100 text-blue-600', emerald: 'bg-emerald-100 text-emerald-600', red: 'bg-red-100 text-red-600', gray: 'bg-gray-100 text-gray-600' } as any)[color];

  const Bar = ({ value, total, color = 'bg-indigo-500' }: { value: number; total: number; color?: string }) => {
    const pct = total > 0 ? Math.round((value / total) * 100) : 0;
    return (
      <div className="w-full bg-gray-100 rounded-full h-3">
        <div className={`${color} h-3 rounded-full`} style={{ width: `${pct}%` }} />
      </div>
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500">Overview of your support activities</p>
      </div>

      {/* Conversation Stats */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Conversations</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard icon={MessageSquare} label="Total Conversations" value={convCounts.total} color="indigo" />
          <StatCard icon={Clock} label="Open" value={convCounts.open} color="amber" />
          <StatCard icon={Users} label="Assigned" value={convCounts.assigned} color="blue" />
          <StatCard icon={CheckCircle} label="Closed" value={convCounts.closed} color="emerald" />
        </div>
        <div className="mt-4 bg-white rounded-xl shadow-sm p-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Conversations distribution</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">Open</div>
              <div className="text-sm font-medium text-gray-900">{convCounts.open}</div>
            </div>
            <Bar value={convCounts.open} total={Math.max(1, convCounts.total)} color="bg-amber-500" />
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">Assigned</div>
              <div className="text-sm font-medium text-gray-900">{convCounts.assigned}</div>
            </div>
            <Bar value={convCounts.assigned} total={Math.max(1, convCounts.total)} color="bg-blue-500" />
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">Closed</div>
              <div className="text-sm font-medium text-gray-900">{convCounts.closed}</div>
            </div>
            <Bar value={convCounts.closed} total={Math.max(1, convCounts.total)} color="bg-emerald-500" />
          </div>
        </div>
      </div>

      {/* Ticket Stats */}
      <div>
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Tickets</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
          <StatCard icon={Ticket} label="Total Tickets" value={ticketCounts.total} color="indigo" />
          <StatCard icon={AlertCircle} label="Open" value={ticketCounts.open} color="red" />
          <StatCard icon={TrendingUp} label="In Progress" value={ticketCounts.in_progress} color="amber" />
          <StatCard icon={CheckCircle} label="Resolved" value={ticketCounts.resolved} color="emerald" />
          <StatCard icon={CheckCircle} label="Closed" value={ticketCounts.closed} color="gray" />
        </div>
        <div className="mt-4 bg-white rounded-xl shadow-sm p-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Tickets distribution</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">Open</div>
              <div className="text-sm font-medium text-gray-900">{ticketCounts.open}</div>
            </div>
            <Bar value={ticketCounts.open} total={Math.max(1, ticketCounts.total)} color="bg-red-500" />
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">In Progress</div>
              <div className="text-sm font-medium text-gray-900">{ticketCounts.in_progress}</div>
            </div>
            <Bar value={ticketCounts.in_progress} total={Math.max(1, ticketCounts.total)} color="bg-amber-500" />
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">Resolved</div>
              <div className="text-sm font-medium text-gray-900">{ticketCounts.resolved}</div>
            </div>
            <Bar value={ticketCounts.resolved} total={Math.max(1, ticketCounts.total)} color="bg-emerald-500" />
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600">Closed</div>
              <div className="text-sm font-medium text-gray-900">{ticketCounts.closed}</div>
            </div>
            <Bar value={ticketCounts.closed} total={Math.max(1, ticketCounts.total)} color="bg-gray-500" />
          </div>
        </div>
      </div>

      {/* Recent Activity (static fallback) */}
      <div className="bg-white rounded-xl shadow-sm p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">Recent Activity</h2>
        <div className="space-y-4">
          {[
            { action: 'New conversation created', channel: 'WhatsApp', time: '2 minutes ago' },
            { action: 'Ticket escalated', channel: 'Instagram', time: '15 minutes ago' },
            { action: 'Conversation assigned to Agent', channel: 'Email', time: '1 hour ago' },
            { action: 'Ticket resolved', channel: 'WhatsApp', time: '2 hours ago' },
            { action: 'New customer message', channel: 'Telegram', time: '3 hours ago' },
          ].map((activity, index) => (
            <div key={index} className="flex items-center gap-4 p-3 bg-gray-50 rounded-lg">
              <div className="w-2 h-2 bg-indigo-500 rounded-full" />
              <div className="flex-1">
                <p className="text-gray-900">{activity.action}</p>
                <p className="text-sm text-gray-500">Channel: {activity.channel}</p>
              </div>
              <span className="text-sm text-gray-400">{activity.time}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

interface StatCardProps {
  icon: React.ElementType;
  label: string;
  value: number;
  color: 'indigo' | 'amber' | 'blue' | 'emerald' | 'red' | 'gray';
}

function StatCard({ icon: Icon, label, value, color }: StatCardProps) {
  const colors = {
    indigo: 'bg-indigo-100 text-indigo-600',
    amber: 'bg-amber-100 text-amber-600',
    blue: 'bg-blue-100 text-blue-600',
    emerald: 'bg-emerald-100 text-emerald-600',
    red: 'bg-red-100 text-red-600',
    gray: 'bg-gray-100 text-gray-600',
  };

  return (
    <div className="bg-white rounded-xl shadow-sm p-6">
      <div className="flex items-center gap-4">
        <div className={`p-3 rounded-lg ${colors[color]}`}>
          <Icon className="w-6 h-6" />
        </div>
        <div>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
          <p className="text-sm text-gray-500">{label}</p>
        </div>
      </div>
    </div>
  );
}
