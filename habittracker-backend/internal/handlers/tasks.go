package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"habittracker/internal/models"

	"github.com/gin-gonic/gin"
)

// GetTasks возвращает задачи пользователя на сегодня
func GetTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		now := time.Now()
		today := now.Format("2006-01-02")
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

		//today := "2025-12-26"
		//yesterday := "2025-12-25"

		fmt.Printf("GetTasks: user=%d, today=%s, yesterday=%s\n", userID, today, yesterday)

		// 1. Проверяем есть ли задачи на сегодня
		var todayCount int
		db.QueryRow(`
            SELECT COUNT(*) 
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            WHERE ut.user_id = ? AND tp.date = ?
        `, userID, today).Scan(&todayCount)

		if todayCount == 0 {
			fmt.Printf("Создаем задачи на %s для user %d\n", today, userID)

			// Убрали проверку на ut.is_active = 1
			_, err := db.Exec(`
                INSERT INTO task_progress (user_task_id, date, completed, streak_days, total_reward)
                SELECT 
                    ut.id, 
                    ?, 
                    0, 
                    CASE 
                        -- Если вчера задача существовала И была выполнена
                        WHEN EXISTS (
                            SELECT 1 
                            FROM task_progress tp_yest
                            WHERE tp_yest.user_task_id = ut.id 
                              AND tp_yest.date = ? 
                              AND tp_yest.completed = 1
                        ) THEN (
                            -- Берем вчерашний streak
                            SELECT tp_yest.streak_days
                            FROM task_progress tp_yest
                            WHERE tp_yest.user_task_id = ut.id 
                              AND tp_yest.date = ? 
                              AND tp_yest.completed = 1
                        )
                        -- Если вчера не выполнена или нет записи - СБРАСЫВАЕМ в 0
                        ELSE 0
                    END,
                    0
                FROM user_tasks ut
                WHERE ut.user_id = ?  -- Убрали: AND ut.is_active = 1
            `, today, yesterday, yesterday, userID)

			if err != nil {
				fmt.Printf("Ошибка создания задач: %v\n", err)
			} else {
				fmt.Printf("Задачи созданы с правильным streak\n")
			}
		}

		// 2. Получаем задачи на сегодня
		rows, err := db.Query(`
            SELECT 
                tp.id,
                tp.user_task_id,
                tp.date,
                tp.completed,
                tp.streak_days,  
                tp.total_reward,
                tt.description,
                tt.id as template_id
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            JOIN task_templates tt ON ut.template_id = tt.id
            WHERE ut.user_id = ? AND tp.date = ?
            ORDER BY tp.id
        `, userID, today)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения задач: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		type TaskResponse struct {
			ID          int    `json:"id"`
			UserTaskID  int    `json:"user_task_id"`
			Date        string `json:"date"`
			Completed   bool   `json:"completed"`
			StreakDays  int    `json:"streak_days"`
			TotalReward int    `json:"total_reward"`
			Description string `json:"description"`
			TemplateID  int    `json:"template_id"`
		}

		var tasks []TaskResponse
		for rows.Next() {
			var task TaskResponse
			var dateStr string

			rows.Scan(&task.ID, &task.UserTaskID, &dateStr,
				&task.Completed, &task.StreakDays, &task.TotalReward,
				&task.Description, &task.TemplateID)

			task.Date = dateStr[:10]
			tasks = append(tasks, task)

			fmt.Printf("Задача %d: streak_days=%d, completed=%v\n",
				task.ID, task.StreakDays, task.Completed)
		}

		c.JSON(http.StatusOK, gin.H{
			"date":  today,
			"tasks": tasks,
			"total": len(tasks),
		})
	}
}

// SubmitTasks отправляет выполненные задачи
func SubmitTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		now := time.Now()
		today := now.Format("2006-01-02")
		//today := "2025-12-26"

		fmt.Printf("SubmitTasks: user=%d, today=%s\n", userID, today)

		var req models.TaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		if len(req.TaskIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нет задач для отправки"})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала транзакции"})
			return
		}

		totalReward := 0
		tasksCompleted := 0

		for _, taskID := range req.TaskIDs {
			// Получаем текущий streak (уже правильный из GetTasks) и base_reward
			var currentStreak, baseReward int
			err := tx.QueryRow(`
                SELECT tp.streak_days, tt.base_reward
                FROM task_progress tp
                JOIN user_tasks ut ON tp.user_task_id = ut.id
                JOIN task_templates tt ON ut.template_id = tt.id
                WHERE tp.id = ? AND ut.user_id = ? AND tp.date = ? AND tp.completed = 0
            `, taskID, userID, today).Scan(&currentStreak, &baseReward)

			if err != nil {
				fmt.Printf("Задача %d не найдена или уже выполнена: %v\n", taskID, err)
				continue
			}

			fmt.Printf("Задача %d: текущий streak=%d\n", taskID, currentStreak)

			// ВАЖНОЕ ИЗМЕНЕНИЕ:
			// Если currentStreak = 0 (значит вчера задача не была выполнена),
			// то это ПЕРВОЕ выполнение в новой серии - streak должен стать 1
			// Если currentStreak > 0 (вчера была выполнена), то увеличиваем на 1

			newStreak := 1 // По умолчанию начинаем с 1
			if currentStreak > 0 {
				newStreak = currentStreak + 1
			}

			fmt.Printf("Текущий streak=%d, новый streak=%d\n", currentStreak, newStreak)

			// РАСЧЕТ НАГРАДЫ
			reward := baseReward * newStreak

			fmt.Printf("Награда=%d\n", reward)

			// ОБНОВЛЯЕМ ЗАДАЧУ - УСТАНАВЛИВАЕМ НОВЫЙ STREAK
			_, err = tx.Exec(`
                UPDATE task_progress 
                SET completed = 1, 
                    total_reward = ?,
                    streak_days = ?  
                WHERE id = ? AND user_task_id IN (
                    SELECT id FROM user_tasks WHERE user_id = ?
                )
            `, reward, newStreak, taskID, userID)

			if err != nil {
				fmt.Printf("Ошибка обновления задачи %d: %v\n", taskID, err)
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка обновления задачи: " + err.Error(),
				})
				return
			}

			totalReward += reward
			tasksCompleted++
		}

		if tasksCompleted == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Не удалось обновить задачи",
				"hint":  "Возможно задачи уже выполнены или не принадлежат вам",
			})
			return
		}

		// Начисляем монеты
		_, err = tx.Exec(`UPDATE users SET coins = coins + ? WHERE id = ?`, totalReward, userID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начисления монет"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка завершения операции"})
			return
		}

		var newBalance int
		db.QueryRow("SELECT coins FROM users WHERE id = ?", userID).Scan(&newBalance)

		c.JSON(http.StatusOK, gin.H{
			"message":         "Задачи отправлены",
			"tasks_completed": tasksCompleted,
			"coins_earned":    totalReward,
			"new_balance":     newBalance,
		})
	}
}

// GetUserStats остается без изменений
func GetUserStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		// Общее количество выполненных задач
		var totalCompleted int
		db.QueryRow(`
            SELECT COUNT(*) 
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            WHERE ut.user_id = ? AND tp.completed = 1
        `, userID).Scan(&totalCompleted)

		// Выполнено за последние 7 дней
		var weeklyCompleted int
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		db.QueryRow(`
            SELECT COUNT(*) 
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            WHERE ut.user_id = ? AND tp.completed = 1 AND tp.date >= ?
        `, userID, weekAgo).Scan(&weeklyCompleted)

		// Текущий баланс
		var coins int
		db.QueryRow("SELECT coins FROM users WHERE id = ?", userID).Scan(&coins)

		// Общая награда
		var totalReward int
		db.QueryRow(`
            SELECT COALESCE(SUM(tp.total_reward), 0)
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            WHERE ut.user_id = ? AND tp.completed = 1
        `, userID).Scan(&totalReward)

		// Текущая серия дней (берем минимальный streak среди ВСЕХ задач, а не только активных)
		var currentStreak int
		db.QueryRow(`
            SELECT COALESCE(MIN(tp.streak_days), 0)
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            WHERE ut.user_id = ? AND tp.date = (
                SELECT MAX(date) FROM task_progress tp2
                JOIN user_tasks ut2 ON tp2.user_task_id = ut2.id
                WHERE ut2.user_id = ?
            )
        `, userID, userID).Scan(&currentStreak)

		// Максимальная серия
		var maxStreak int
		db.QueryRow(`
            SELECT COALESCE(MAX(streak_days), 0)
            FROM task_progress tp
            JOIN user_tasks ut ON tp.user_task_id = ut.id
            WHERE ut.user_id = ?
        `, userID).Scan(&maxStreak)

		c.JSON(http.StatusOK, gin.H{
			"total_completed":     totalCompleted,
			"weekly_completed":    weeklyCompleted,
			"current_coins":       coins,
			"total_reward_earned": totalReward,
			"current_streak":      currentStreak,
			"max_streak":          maxStreak,
		})
	}
}
