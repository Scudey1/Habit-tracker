package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetTaskCatalog возвращает каталог всех доступных задач
func GetTaskCatalog(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		rows, err := db.Query(`
            SELECT 
                tt.id,
                tt.description,
                tt.base_reward,
                CASE WHEN ut.id IS NOT NULL THEN 1 ELSE 0 END as is_selected
            FROM task_templates tt
            LEFT JOIN user_tasks ut ON tt.id = ut.template_id AND ut.user_id = ?
            ORDER BY tt.created_at
        `, userID)

		if err != nil {
			log.Printf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения каталога: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		// Элемент каталога
		type CatalogItem struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
			BaseReward  int    `json:"base_reward"`
			IsSelected  bool   `json:"is_selected"`
		}

		var catalog []CatalogItem
		for rows.Next() {
			var item CatalogItem
			var isSelected int
			rows.Scan(&item.ID, &item.Description, &item.BaseReward, &isSelected)
			item.IsSelected = isSelected == 1
			catalog = append(catalog, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"catalog": catalog,
			"total":   len(catalog),
		})
	}
}

func SelectTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		var req struct {
			TemplateID int `json:"template_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректные данные",
			})
			return
		}

		// Добавляем в user_tasks
		result, err := db.Exec(`
            INSERT OR IGNORE INTO user_tasks (user_id, template_id) 
            VALUES (?, ?)
        `, userID, req.TemplateID)

		if err != nil {
			log.Printf("Insert error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка добавления задачи: " + err.Error(),
			})
			return
		}

		// Получаем ID вставленной записи
		userTaskID, _ := result.LastInsertId()

		if userTaskID > 0 {
			// Создаем task_progress на сегодня
			today := time.Now().Format("2006-01-02")
			_, err = db.Exec(`
                INSERT OR IGNORE INTO task_progress 
                (user_task_id, date, completed, streak_days, total_reward) 
                VALUES (?, ?, 0, 0, 0)
            `, userTaskID, today)

			if err != nil {
				log.Printf("Task progress error: %v", err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Задача добавлена",
		})
	}
}
