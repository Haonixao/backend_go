package scheduled_models

import (
	"github.com/go-co-op/gocron"
)

type Task struct {
	ID          string      `gorm:"primaryKey;column:id;type:text;not null" json:"id"`
	Schedule    string      `gorm:"column:schedule;type:text;not null" json:"schedule"`
	Description string      `gorm:"column:description;type:text;not null" json:"description"`
	IsEnabled   bool        `gorm:"column:is_enabled;type:boolean;not null" json:"is_enabled"`
	TaskFunc    interface{} `gorm:"-" json:"-"`
	CronJob     *gocron.Job `gorm:"-" json:"-"`
}

func (t *Task) TableName() string {
	return "scheduled.task"
}

func (t *Task) Hash() string {
	return t.ID
}
