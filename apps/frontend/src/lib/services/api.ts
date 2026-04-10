const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

interface RequestOptions {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, headers = {} } = options;

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...headers
    },
    body: body ? JSON.stringify(body) : undefined
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  return response.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string, headers?: Record<string, string>) =>
    request<T>(path, { method: 'GET', headers }),

  post: <T>(path: string, body: unknown, headers?: Record<string, string>) =>
    request<T>(path, { method: 'POST', body, headers })
};

export async function fetchHealthcheck(): Promise<{ status: string }> {
  return api.get<{ status: string }>('/health');
}

export interface BackendProject {
  id: string;
  name: string;
  description: string;
  tech: string[];
  url?: string;
}

export async function fetchProjects(): Promise<BackendProject[]> {
  return api.get<BackendProject[]>('/projects');
}
