export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'agent';
  tenant_id: string;
  created_at: string;
  updated_at: string;
}

export interface Conversation {
  id: string;
  tenant_id: string;
  customer_id: string;
  customer_name: string;
  customer_external_id: string;
  channel: string;
  status: 'open' | 'assigned' | 'closed';
  assigned_agent_id: string | null;
  assigned_agent_name?: string;
  last_message?: string;
  last_message_at?: string;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_type: 'customer' | 'agent';
  sender_name: string;
  message: string;
  created_at: string;
}

export interface Ticket {
  id: string;
  tenant_id: string;
  conversation_id?: string | null;
  code?: string | null;
  title: string;
  description: string;
  status: 'open' | 'in_progress' | 'resolved' | 'closed';
  priority: 'stabdar' | 'regular' | 'VIP';
  assigned_agent_id: string | null;
  assigned_agent_name?: string;
  created_by_id: string;
  created_by_name?: string;
  created_at: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export interface DashboardStats {
  total_conversations: number;
  open_conversations: number;
  assigned_conversations: number;
  closed_conversations: number;
  total_tickets: number;
  open_tickets: number;
  in_progress_tickets: number;
  resolved_tickets: number;
}
