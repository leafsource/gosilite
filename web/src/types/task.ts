export type TaskStatus = 'todo' | 'doing' | 'done'

export interface Task {
  id: string
  title: string
  description: string
  status: TaskStatus
  created_at: string
  updated_at: string
}

export interface CreateTaskRequest {
  title: string
  description?: string
  status?: TaskStatus
}

export interface UpdateTaskRequest {
  title?: string
  description?: string
  status?: TaskStatus
}

export interface ApiEnvelope<T> {
  ok: boolean
  data?: T
  error?: {
    code: string
    message: string
    detail?: Record<string, unknown>
  }
}
