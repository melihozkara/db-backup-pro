// Detect whether we're running in Wails (desktop) or Web (HTTP) mode

let cachedMode: 'wails' | 'web' | null = null;

export function getAppMode(): 'wails' | 'web' {
  if (cachedMode) return cachedMode;

  // Check if Wails runtime is available
  const w = window as any;
  if (w.go?.main?.App) {
    cachedMode = 'wails';
  } else {
    cachedMode = 'web';
  }

  return cachedMode;
}

export function isWailsMode(): boolean {
  return getAppMode() === 'wails';
}

export function isWebMode(): boolean {
  return getAppMode() === 'web';
}
