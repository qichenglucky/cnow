import client from './client';
import type { ApiResponse, PaginatedResponse, Release, CreateReleaseInput } from '../types/api';

export async function fetchReleases(params: {
  service_id?: string;
  status?: string;
  offset?: number;
  limit?: number;
}) {
  const res = await client.get<ApiResponse<PaginatedResponse<Release>>>('/releases', { params });
  return res.data.data;
}

export async function fetchRelease(id: string) {
  const res = await client.get<ApiResponse<Release>>(`/releases/${id}`);
  return res.data.data;
}

export async function createRelease(input: CreateReleaseInput) {
  const res = await client.post<ApiResponse<Release>>('/releases', input);
  return res.data.data;
}
