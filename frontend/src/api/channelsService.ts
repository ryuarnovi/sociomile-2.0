import client from './client';

export type ChannelPayload = {
  name: string;
  slug?: string;
  description?: string;
};

export const listChannels = async () => {
  const res = await client.get('/channels');
  return res.data?.data ?? res.data;
};

export const getChannel = async (id: string) => {
  const res = await client.get(`/channels/${id}`);
  return res.data?.data ?? res.data;
};

export const createChannel = async (payload: ChannelPayload) => {
  const res = await client.post('/channels', payload);
  return res.data?.data ?? res.data;
};

export const updateChannel = async (id: string, payload: Partial<ChannelPayload>) => {
  const res = await client.put(`/channels/${id}`, payload);
  return res.data?.data ?? res.data;
};

export const deleteChannel = async (id: string) => {
  const res = await client.delete(`/channels/${id}`);
  return res.data?.success ?? res.data;
};
