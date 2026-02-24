import client from './client';

export type ConversationCreatePayload = { title?: string; participants?: number[] };

export const listConversations = async () => {
  const res = await client.get('/conversations');
  return res.data?.data ?? res.data;
};

export const getConversation = async (id: string | number) => {
  const res = await client.get(`/conversations/${id}`);
  return res.data?.data ?? res.data;
};

export const createConversation = async (payload: ConversationCreatePayload) => {
  const res = await client.post('/conversations', payload);
  return res.data;
};

export const updateConversation = async (id: number, payload: Partial<ConversationCreatePayload>) => {
  const res = await client.put(`/conversations/${id}`, payload);
  return res.data;
};

export const deleteConversation = async (id: number) => {
  const res = await client.delete(`/conversations/${id}`);
  return res.data;
};

export const listTicketsForConversation = async (conversationId: string) => {
  const res = await client.get(`/conversations/${conversationId}/tickets`);
  return res.data?.data ?? res.data;
};

export const setSelectedTicket = async (conversationId: string, ticketId: string) => {
  const res = await client.put(`/conversations/${conversationId}/selected-ticket`, { ticket_id: ticketId });
  return res.data;
};

export const assignConversation = async (conversationId: string) => {
  const res = await client.post(`/conversations/${conversationId}/assign`);
  return res.data;
};

export const closeConversation = async (conversationId: string) => {
  const res = await client.post(`/conversations/${conversationId}/close`);
  return res.data;
};
