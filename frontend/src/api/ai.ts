import client from './client';
import type { ApiResponse, AIPlan, RiskAnalysis } from '../types/api';

export async function requestAIPlan(prompt: string) {
  const res = await client.post<ApiResponse<AIPlan>>('/ai/plan', { prompt });
  return res.data.data;
}

export async function requestRiskAnalysis(payload: { service_id: string; environment: string; version: string; prompt?: string }) {
  const res = await client.post<ApiResponse<RiskAnalysis>>('/ai/risk', payload);
  return res.data.data;
}
