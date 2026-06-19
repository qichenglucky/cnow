import { useQuery } from '@tanstack/react-query';
import { fetchLogs, fetchMetrics, fetchAlerts } from '../api/observability';
import { mockLatencyData, mockErrorRateData, mockAlerts, mockLogs } from '../mock/data';
import type { LogsResponse, MetricsResponse, AlertsResponse } from '../api/observability';

export function useObsLogs(params?: { service_id?: string; level?: string; limit?: number }) {
  return useQuery<LogsResponse>({
    queryKey: ['obs-logs', params],
    queryFn: async () => {
      try {
        return await fetchLogs(params);
      } catch {
        return { items: mockLogs, total: mockLogs.length };
      }
    },
  });
}

export function useObsMetrics(params?: { service_id?: string; period?: string }) {
  return useQuery<MetricsResponse>({
    queryKey: ['obs-metrics', params],
    queryFn: async () => {
      try {
        return await fetchMetrics(params);
      } catch {
        return {
          latency: mockLatencyData,
          error_rate: mockErrorRateData,
          throughput: mockLatencyData.map((d) => ({ time: d.time, requests: 1000, success: 950 })),
        };
      }
    },
  });
}

export function useObsAlerts(params?: { service_id?: string }) {
  return useQuery<AlertsResponse>({
    queryKey: ['obs-alerts', params],
    queryFn: async () => {
      try {
        return await fetchAlerts(params);
      } catch {
        return { items: mockAlerts, total: mockAlerts.length };
      }
    },
  });
}
