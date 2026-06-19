import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchServices, fetchService, createService, updateService, deleteService, fetchEnvironments } from '../api/services';
import { mockServices, mockEnvironments } from '../mock/data';
import type { CreateServiceInput } from '../types/api';

export function useServices(offset = 0, limit = 20, search?: string) {
  return useQuery({
    queryKey: ['services', offset, limit, search],
    queryFn: async () => {
      try {
        return await fetchServices(offset, limit, search);
      } catch {
        const filtered = search
          ? mockServices.filter((s) => s.name.includes(search) || s.description.includes(search))
          : mockServices;
        return {
          items: filtered.slice(offset, offset + limit),
          total: filtered.length,
          offset,
          limit,
        };
      }
    },
  });
}

export function useService(id: string) {
  return useQuery({
    queryKey: ['service', id],
    queryFn: async () => {
      try {
        return await fetchService(id);
      } catch {
        return mockServices.find((s) => s.id === id) ?? mockServices[0];
      }
    },
  });
}

export function useCreateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateServiceInput) => createService(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}

export function useUpdateService(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: Partial<CreateServiceInput>) => updateService(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['services'] });
      qc.invalidateQueries({ queryKey: ['service', id] });
    },
  });
}

export function useDeleteService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteService(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}

export function useEnvironments(serviceId: string) {
  return useQuery({
    queryKey: ['environments', serviceId],
    queryFn: async () => {
      try {
        return await fetchEnvironments(serviceId);
      } catch {
        return mockEnvironments.filter((e) => e.service_id === serviceId);
      }
    },
  });
}
