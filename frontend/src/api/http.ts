// HTTP client wrapper for web mode API calls

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = { method };

  if (body !== undefined) {
    opts.headers = { 'Content-Type': 'application/json' };
    opts.body = JSON.stringify(body);
  }

  const res = await fetch(path, opts);

  if (!res.ok) {
    let msg = res.statusText;
    try {
      const err = await res.json();
      if (err.error) msg = err.error;
    } catch {
      // ignore parse errors
    }
    throw new ApiError(msg, res.status);
  }

  // Handle empty responses
  const text = await res.text();
  if (!text) return {} as T;
  return JSON.parse(text) as T;
}

export const http = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: <T>(path: string) => request<T>('DELETE', path),
};
