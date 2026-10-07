package tasks

type UpdateTaskRequest struct {
	Schedule  *string `form:"schedule" json:"schedule" `
	IsEnabled *bool   `form:"is_enabled" json:"is_enabled"`
}
