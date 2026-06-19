import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchReleases, fetchRelease, createRelease } from '../api/releases';
import { mockReleases } from '../mock/data';
import type { CreateReleaseInput } from '../types/api';

export function useReleases(params: {
  service_id?: string;
  status?: string;
  offset?: number;
  limit?: number;
}) {
  return useQuery({
    queryKey: ['releases', params],
    queryFn: async () => {
      try {
        return await fetchReleases(params);
      } catch {
        let filtered = mockReleases;
        if (params.service_id) filtered = filtered.filter((r) => r.service_id === params.service_id);
        if (params.status) filtered = filtered.filter((r) => r.status === params.status);
        const offset = params.offset ?? 0;
        const limit = params.limit ?? 20;
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

export function useRelease(id: string) {
  return useQuery({
    queryKey: ['release', id],
    queryFn: async () => {
      try {
        return await fetchRelease(id);
      } catch {
        return mockReleases.find((r) => r.id === id) ?? mockReleases[0];
      }
    },
  });
}

export function useCreateRelease() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateReleaseInput) => createRelease(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['releases'] }),
  });
}
