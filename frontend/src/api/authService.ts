import client from './client';

export type LoginPayload = { email: string; password: string };
export type RegisterPayload = { email: string; password: string; name?: string };

export const login = async (payload: LoginPayload) => {
  const res = await client.post('/auth/login', payload);
  return res.data;
};

export const register = async (payload: RegisterPayload) => {
  const res = await client.post('/auth/register', payload);
  return res.data;
};

export const logout = async () => {
  // implement if backend supports logout endpoint
  return true;
};
