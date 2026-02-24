import client from './client';

export type MessagePayload = { conversationId: number | string; message: string };

export const listMessages = async (conversationId: number | string) => {
  // backend exposes messages as part of conversation GET
  const res = await client.get(`/conversations/${conversationId}`);
  const data = res.data?.data ?? res.data;
  return data?.messages ?? [];
};

export const createMessage = async (conversationId: number | string, message: string) => {
  const res = await client.post(`/conversations/${conversationId}/messages`, { message });
  return res.data?.data ?? res.data;
};

export const deleteMessage = async (id: number | string) => {
  const res = await client.delete(`/messages/${id}`);
  return res.data;
};
