package scheduled_models

import "time"

type TaskLock struct {
	TaskID    string    `json:"task_id" gorm:"primaryKey;column:task_id;type:text;not null"`
	Task      *Task     `json:"-" gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:timestamp with time zone;not null"`
}

func (TaskLock) TableName() string {
	return "scheduled.task_lock"
}
