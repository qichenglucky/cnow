import axios from 'axios';
import { message } from 'antd';
import type { ApiResponse } from '../types/api';

let requestCounter = 0;

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
});

// Request interceptor
client.interceptors.request.use((config) => {
  const reqId = `req-${Date.now()}-${++requestCounter}`;
  config.headers['X-Request-Id'] = reqId;
  // Auth headers (MVP: read from localStorage, default to admin)
  config.headers['X-User-Id'] = localStorage.getItem('x-user-id') || '1';
  config.headers['X-User-Role'] = localStorage.getItem('x-user-role') || 'admin';
  if (config.method === 'post') {
    config.headers['X-Idempotency-Key'] = `idem-${Date.now()}-${requestCounter}`;
  }
  return config;
});

// Response interceptor
client.interceptors.response.use(
  (response) => {
    const envelope = response.data as ApiResponse<unknown>;
    if (envelope && envelope.code !== undefined && envelope.code !== 0) {
      message.error(envelope.message || '请求失败');
      return Promise.reject(new Error(envelope.message));
    }
    return response;
  },
  (error) => {
    const msg = error.response?.data?.message || error.message || '网络错误';
    message.error(msg);
    return Promise.reject(error);
  },
);

export default client;
