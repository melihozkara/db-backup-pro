// Dual-mode event system: Wails runtime events in desktop mode, SSE in web mode

import { isWailsMode } from './mode';

type EventCallback = (data: any) => void;

// SSE connection singleton for web mode
let eventSource: EventSource | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
const webListeners = new Map<string, Set<EventCallback>>();
// Track which event names are already registered on the current EventSource
const registeredSSEEvents = new Set<string>();

// Wails mode: listener registry to avoid EventsOff removing all listeners for an event
const wailsListeners = new Map<string, Set<EventCallback>>();

function ensureSSE() {
  if (eventSource) return;

  const connect = () => {
    registeredSSEEvents.clear();
    eventSource = new EventSource('/api/events');

    eventSource.onopen = () => {
      console.log('[SSE] Connected');
    };

    eventSource.onerror = () => {
      console.log('[SSE] Connection lost, reconnecting in 3s...');
      eventSource?.close();
      eventSource = null;
      registeredSSEEvents.clear();
      if (!reconnectTimer) {
        reconnectTimer = setTimeout(() => {
          reconnectTimer = null;
          if (webListeners.size > 0) connect();
        }, 3000);
      }
    };

    // Register all pending event types on this new EventSource
    for (const eventName of webListeners.keys()) {
      addSSEListener(eventName);
    }
  };

  connect();
}

function addSSEListener(eventName: string) {
  if (!eventSource || registeredSSEEvents.has(eventName)) return;
  registeredSSEEvents.add(eventName);

  eventSource.addEventListener(eventName, (e: MessageEvent) => {
    try {
      const data = JSON.parse(e.data);
      const callbacks = webListeners.get(eventName);
      if (callbacks) {
        callbacks.forEach((cb) => cb(data));
      }
    } catch (err) {
      console.error('[SSE] Parse error:', err);
    }
  });
}

/**
 * Subscribe to an event. Works in both Wails and Web mode.
 */
export function EventsOn(eventName: string, callback: EventCallback): void {
  if (isWailsMode()) {
    // Track callbacks in our registry
    if (!wailsListeners.has(eventName)) {
      wailsListeners.set(eventName, new Set());
    }
    wailsListeners.get(eventName)!.add(callback);

    // If this is the first listener for this event, register a dispatcher with Wails
    if (wailsListeners.get(eventName)!.size === 1) {
      import('../../wailsjs/runtime/runtime').then((mod) => {
        mod.EventsOn(eventName, (data: any) => {
          const callbacks = wailsListeners.get(eventName);
          if (callbacks) {
            callbacks.forEach((cb) => cb(data));
          }
        });
      });
    }
    return;
  }

  // Web mode: use SSE
  if (!webListeners.has(eventName)) {
    webListeners.set(eventName, new Set());
  }
  webListeners.get(eventName)!.add(callback);
  ensureSSE();

  // Always register the SSE listener (addSSEListener is idempotent)
  if (eventSource) {
    addSSEListener(eventName);
  }
}

/**
 * Unsubscribe from an event.
 */
export function EventsOff(eventName: string, callback?: EventCallback): void {
  if (isWailsMode()) {
    if (callback) {
      wailsListeners.get(eventName)?.delete(callback);
    } else {
      wailsListeners.delete(eventName);
    }
    // Only call Wails EventsOff when no listeners remain for this event
    if (!wailsListeners.get(eventName)?.size) {
      wailsListeners.delete(eventName);
      import('../../wailsjs/runtime/runtime').then((mod) => {
        mod.EventsOff(eventName);
      });
    }
    return;
  }

  // Web mode
  if (callback) {
    webListeners.get(eventName)?.delete(callback);
    if (webListeners.get(eventName)?.size === 0) {
      webListeners.delete(eventName);
    }
  } else {
    webListeners.delete(eventName);
  }

  // Close SSE if no listeners remain
  if (webListeners.size === 0 && eventSource) {
    eventSource.close();
    eventSource = null;
    registeredSSEEvents.clear();
  }
}
