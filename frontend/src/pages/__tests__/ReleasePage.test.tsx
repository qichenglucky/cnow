import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import ReleasePage from '../ReleasePage';

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

describe('ReleasePage', () => {
  it('renders release list with mock data', async () => {
    renderWithProviders(<ReleasePage />);

    // Wait for mock data to load (API will fail, fallback to mock)
    await waitFor(
      () => {
        expect(screen.getByText('v1.1.0')).toBeInTheDocument();
      },
      { timeout: 5000 },
    );
  });

  it('shows status badges in the table', async () => {
    renderWithProviders(<ReleasePage />);

    // Wait for data to load, then check that StatusBadge components render
    await waitFor(
      () => {
        // The first release has status 'created' which renders '已创建'
        expect(screen.getByText('已创建')).toBeInTheDocument();
      },
      { timeout: 5000 },
    );
  });

  it('has a create release button', () => {
    renderWithProviders(<ReleasePage />);
    const createBtn = screen.getByRole('button', { name: /创建发布/i });
    expect(createBtn).toBeInTheDocument();
  });
});
