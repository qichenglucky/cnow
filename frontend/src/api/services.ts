import client from './client';
import type { ApiResponse, PaginatedResponse, Service, CreateServiceInput, Environment } from '../types/api';

export async function fetchServices(offset = 0, limit = 20, search?: string) {
  const params: Record<string, string | number> = { offset, limit };
  if (search) params.search = search;
  const res = await client.get<ApiResponse<PaginatedResponse<Service>>>('/services', { params });
  return res.data.data;
}

export async function fetchService(id: string) {
  const res = await client.get<ApiResponse<Service>>(`/services/${id}`);
  return res.data.data;
}

export async function createService(input: CreateServiceInput) {
  const res = await client.post<ApiResponse<Service>>('/services', input);
  return res.data.data;
}

export async function updateService(id: string, input: Partial<CreateServiceInput>) {
  const res = await client.put<ApiResponse<Service>>(`/services/${id}`, input);
  return res.data.data;
}

export async function deleteService(id: string) {
  const res = await client.delete<ApiResponse<null>>(`/services/${id}`);
  return res.data.data;
}

export async function fetchEnvironments(serviceId: string) {
  const res = await client.get<ApiResponse<Environment[]>>('/environments', {
    params: { service_id: serviceId },
  });
  return res.data.data;
}

export async function createEnvironment(input: Partial<Environment>) {
  const res = await client.post<ApiResponse<Environment>>('/environments', input);
  return res.data.data;
}
