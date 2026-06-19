import client from './client';
import type { ApiResponse } from '../types/api';

// ---- Observability types ----

export interface LogEntry {
  id: string;
  timestamp: string;
  level: string;
  service: string;
  message: string;
  traceId: string;
}

export interface LatencyPoint {
  time: string;
  p50: number;
  p95: number;
  p99: number;
}

export interface ErrorRatePoint {
  time: string;
  rate: number;
  count: number;
}

export interface ThroughputPoint {
  time: string;
  requests: number;
  success: number;
}

export interface AlertRule {
  id: string;
  level: string;
  service: string;
  message: string;
  time: string;
  status: string;
}

export interface LogsResponse {
  items: LogEntry[];
  total: number;
}

export interface MetricsResponse {
  latency: LatencyPoint[];
  error_rate: ErrorRatePoint[];
  throughput: ThroughputPoint[];
}

export interface AlertsResponse {
  items: AlertRule[];
  total: number;
}

// ---- API functions ----

export async function fetchLogs(params?: {
  service_id?: string;
  env?: string;
  level?: string;
  limit?: number;
}): Promise<LogsResponse> {
  const res = await client.get<ApiResponse<LogsResponse>>('/observability/logs', { params });
  return res.data.data;
}

export async function fetchMetrics(params?: {
  service_id?: string;
  period?: string;
}): Promise<MetricsResponse> {
  const res = await client.get<ApiResponse<MetricsResponse>>('/observability/metrics', { params });
  return res.data.data;
}

export async function fetchAlerts(params?: {
  service_id?: string;
}): Promise<AlertsResponse> {
  const res = await client.get<ApiResponse<AlertsResponse>>('/observability/alerts', { params });
  return res.data.data;
}
