import client from './client';

export type UserPayload = {
  name?: string;
  email?: string;
  password?: string;
  role?: string;
};

export const listUsers = async () => {
  const res = await client.get('/users');
  return res.data;
};

export const getUser = async (id: string) => {
  const res = await client.get(`/users/${id}`);
  return res.data;
};

export const createUser = async (payload: UserPayload) => {
  const res = await client.post('/users', payload);
  return res.data;
};

export const updateUser = async (id: string, payload: Partial<UserPayload>) => {
  const res = await client.put(`/users/${id}`, payload);
  return res.data;
};

export const deleteUser = async (id: string) => {
  const res = await client.delete(`/users/${id}`);
  return res.data;
};
