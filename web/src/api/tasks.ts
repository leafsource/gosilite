import type {
  ApiEnvelope,
  CreateTaskRequest,
  Task,
  UpdateTaskRequest,
} from '../types/task'

export class ApiError extends Error {
  code: string
  constructor(code: string, message: string) {
    super(message)
    this.code = code
    this.name = 'ApiError'
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })
  if (res.status === 204) {
    return undefined as T
  }
  const body = (await res.json()) as ApiEnvelope<T>
  if (!body.ok) {
    throw new ApiError(body.error?.code ?? 'internal', body.error?.message ?? res.statusText)
  }
  return body.data as T
}

export function listTasks(status?: string): Promise<Task[]> {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return request<Task[]>(`/api/tasks${query}`)
}

export function getTask(id: string): Promise<Task> {
  return request<Task>(`/api/tasks/${encodeURIComponent(id)}`)
}

export function createTask(input: CreateTaskRequest): Promise<Task> {
  return request<Task>('/api/tasks', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function updateTask(id: string, input: UpdateTaskRequest): Promise<Task> {
  return request<Task>(`/api/tasks/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(input),
  })
}

export function deleteTask(id: string): Promise<void> {
  return request<void>(`/api/tasks/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}
