import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useServices, useCreateService } from '../useServices';
import { mockServices } from '../../mock/data';
import React from 'react';

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('useServices', () => {
  it('useServiceList returns mock data when API fails', async () => {
    const { result } = renderHook(() => useServices(0, 10), {
      wrapper: createWrapper(),
    });

    // Initially loading
    expect(result.current.isLoading).toBe(true);

    // Wait for the query to settle (API fails → mock fallback)
    await waitFor(
      () => {
        expect(result.current.data).toBeDefined();
      },
      { timeout: 5000 },
    );

    expect(result.current.data!.items.length).toBeGreaterThan(0);
    expect(result.current.data!.items[0].name).toBe(mockServices[0].name);
  });

  it('useCreateService mutation exists and has correct shape', () => {
    const { result } = renderHook(() => useCreateService(), {
      wrapper: createWrapper(),
    });

    expect(result.current.mutate).toBeDefined();
    expect(typeof result.current.mutate).toBe('function');
    expect(result.current.isPending).toBe(false);
  });
});
