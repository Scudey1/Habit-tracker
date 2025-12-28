package models

import "time"

// Task - ежедневная задача
type Task struct {
	ID          int       `json:"id"`
	UserTaskID  int       `json:"user_task_id"` // Ссылка на user_tasks
	Date        string    `json:"date"`         // Формат: "YYYY-MM-DD"
	Completed   bool      `json:"completed"`
	StreakDays  int       `json:"streak_days"`  // Текущая серия дней
	TotalReward int       `json:"total_reward"` // Награда за этот день
	CreatedAt   time.Time `json:"created_at"`

	// Дополнительные поля для удобства
	Description string `json:"description"` // Из task_templates
	TemplateID  int    `json:"template_id"` // Из task_templates
}

// TaskRequest для отметки выполнения задач
type TaskRequest struct {
	TaskIDs []int `json:"task_ids"` // ID из task_progress
}

// TaskTemplate - шаблон задачи
type TaskTemplate struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	BaseReward  int       `json:"base_reward"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserTask - выбранная пользователем задача
type UserTask struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	TemplateID int       `json:"template_id"`
	CreatedAt  time.Time `json:"created_at"`
}
