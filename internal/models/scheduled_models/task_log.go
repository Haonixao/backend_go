package scheduled_models

import "time"

type TaskLog struct {
	TaskID       string    `json:"task_id" gorm:"primaryKey;column:task_id;type:text;not null"`
	Task         *Task     `json:"-" gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DateStart    time.Time `json:"date_start" gorm:"primaryKey;column:date_start;type:timestamp with time zone;not null"`
	DateEnd      time.Time `json:"date_end" gorm:"column:date_end;type:timestamp with time zone;not null"`
	IsGood       bool      `json:"is_good" gorm:"column:is_good;type:boolean;not null"`
	ErrorMessage *string   `json:"error_message" gorm:"column:error_message;type:text"`
}

func (TaskLog) TableName() string {
	return "scheduled.task_log"
}
