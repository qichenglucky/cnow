import '@testing-library/jest-dom/vitest';

// Mock window.matchMedia (antd responsive components need this)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }),
});

// Mock window.getComputedStyle (antd needs this for some components)
const originalGetComputedStyle = window.getComputedStyle;
window.getComputedStyle = (elt: Element, pseudoElt?: string | null) => {
  const style = originalGetComputedStyle(elt, pseudoElt);
  return style;
};

// Suppress antd console warnings in tests
const originalError = console.error;
console.error = (...args: unknown[]) => {
  if (
    typeof args[0] === 'string' &&
    (args[0].includes('Warning: `NaN` is an invalid value') ||
      args[0].includes('Warning: Unknown event handler') ||
      args[0].includes('Not wrapped in act'))
  ) {
    return;
  }
  originalError(...args);
};
