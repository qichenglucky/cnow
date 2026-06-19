import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import ServiceListPage from '../ServiceListPage';

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('ServiceListPage', () => {
  it('renders table with correct column headers', async () => {
    renderWithProviders(<ServiceListPage />);

    // Wait for the table to render (antd Table renders column headers even when loading)
    await waitFor(() => {
      expect(screen.getByText('名称')).toBeInTheDocument();
    });
    expect(screen.getByText('状态')).toBeInTheDocument();
    expect(screen.getByText('技术栈')).toBeInTheDocument();
    expect(screen.getByText('负责人')).toBeInTheDocument();
    expect(screen.getByText('创建时间')).toBeInTheDocument();
  });

  it('shows mock data when API is unavailable', async () => {
    renderWithProviders(<ServiceListPage />);

    // API will fail in test env, mock data fallback should kick in
    await waitFor(
      () => {
        expect(screen.getByText('用户中心')).toBeInTheDocument();
      },
      { timeout: 5000 },
    );
    expect(screen.getByText('订单服务')).toBeInTheDocument();
  });

  it('has a create service button', () => {
    renderWithProviders(<ServiceListPage />);
    const createBtn = screen.getByRole('button', { name: /创建服务/i });
    expect(createBtn).toBeInTheDocument();
  });

  it('has a search input', () => {
    renderWithProviders(<ServiceListPage />);
    const searchInput = screen.getByPlaceholderText('搜索服务名称');
    expect(searchInput).toBeInTheDocument();
  });
});
