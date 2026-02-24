import client from './client';
import { Ticket } from '../types';

export type TicketPayload = { title: string; description?: string; assigned_agent_id?: string; conversation_id?: string | null; priority?: string; code?: string | null };

export const listTickets = async (params?: { status?: string; priority?: string; page?: number; per_page?: number }) => {
  // backend validation requires page >= 1 and per_page >= 1, so supply defaults
  const safeParams = {
    page: params?.page ?? 1,
    per_page: params?.per_page ?? 20,
    ...(params?.status ? { status: params.status } : {}),
    ...(params?.priority ? { priority: params.priority } : {}),
  };
  const res = await client.get('/tickets', { params: safeParams });
  return res.data; // expects ApiResponse with data array and meta
};

export const getTicket = async (id: string) => {
  const res = await client.get(`/tickets/${id}`);
  return res.data as { success: boolean; data: Ticket };
};

export const createTicket = async (payload: TicketPayload) => {
  const res = await client.post('/tickets', payload);
  return res.data as { success: boolean; data: Ticket };
};

export type EscalatePayload = { title?: string; description?: string; priority?: string; ticket_code?: string };

export const escalateTicket = async (conversationId: string, payload: EscalatePayload) => {
  const res = await client.post(`/conversations/${conversationId}/escalate`, payload);
  return res.data as { success: boolean; data: Ticket };
};

export const updateTicket = async (id: string, payload: Partial<TicketPayload>) => {
  const res = await client.put(`/tickets/${id}`, payload);
  return res.data as { success: boolean; data: Ticket };
};

export const deleteTicket = async (id: string) => {
  const res = await client.delete(`/tickets/${id}`);
  return res.data as { success: boolean };
};

export const updateTicketStatus = async (id: string, status: string) => {
  const res = await client.put(`/tickets/${id}/status`, { status });
  return res.data as { success: boolean };
};
