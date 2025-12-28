package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SearchUsersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем поисковый запрос
		query := strings.TrimSpace(c.Query("q"))

		if query == "" {
			c.JSON(http.StatusOK, gin.H{
				"users": []interface{}{},
				"count": 0,
			})
			return
		}

		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Минимум 2 символа для поиска",
			})
			return
		}

		// Ищем пользователей (регистронезависимо)
		rows, err := db.Query(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER(username) LIKE LOWER(?)
            ORDER BY username
            LIMIT 20
        `, query+"%")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка поиска: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		type UserResult struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
		}

		var users []UserResult
		for rows.Next() {
			var user UserResult
			if err := rows.Scan(&user.ID, &user.Username); err != nil {
				continue
			}
			users = append(users, user)
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
	}
}
