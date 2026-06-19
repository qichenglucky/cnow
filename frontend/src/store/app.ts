import { create } from 'zustand';

interface AppState {
  currentServiceId: string | null;
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  aiDrawerOpen: boolean;
  setCurrentService: (id: string | null) => void;
  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
  toggleAiDrawer: () => void;
}

export const useAppStore = create<AppState>((set) => ({
  currentServiceId: null,
  sidebarCollapsed: false,
  theme: 'light',
  aiDrawerOpen: false,
  setCurrentService: (id) => set({ currentServiceId: id }),
  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
  setTheme: (theme) => set({ theme }),
  toggleAiDrawer: () => set((s) => ({ aiDrawerOpen: !s.aiDrawerOpen })),
}));
