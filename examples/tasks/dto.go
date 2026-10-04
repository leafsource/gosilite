package tasks

// CreateTaskRequest is the create DTO.
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// UpdateTaskRequest is the partial update DTO. nil fields are left unchanged.
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// ListTasksQuery is the list DTO.
type ListTasksQuery struct {
	Status string `json:"status,omitempty"`
}
