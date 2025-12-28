// handlers/user_room.go
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserRoomHandler - получить комнату пользователя (всю его мебель)
func GetUserRoomHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ID пользователя, чью комнату смотрим
		userID, err := strconv.Atoi(c.Param("user_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректный ID пользователя",
			})
			return
		}

		// 1. Проверяем, существует ли пользователь
		var username string
		err = db.QueryRow(`
            SELECT username 
            FROM users 
            WHERE id = ?
        `, userID).Scan(&username)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Пользователь не найден",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка базы данных",
				})
			}
			return
		}

		// 2. Получаем ВСЮ мебель пользователя (без MAX уровня, просто всё)
		rows, err := db.Query(`
            SELECT 
                ft.id,
                ft.name,
                ft.level,
                ft.price,
                ft.max_level,
                ft.created_at
            FROM user_furniture uf
            JOIN furniture_templates ft ON uf.template_id = ft.id
            WHERE uf.user_id = ?
            ORDER BY ft.name, ft.level
        `, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения мебели: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		// Используем ту же структуру, что и в GetMyFurniture
		type FurnitureItem struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Level     int    `json:"level"`
			Price     int    `json:"price"`
			MaxLevel  int    `json:"max_level"`
			CreatedAt string `json:"created_at"`
		}

		var furniture []FurnitureItem
		var totalValue int
		var itemCount int

		for rows.Next() {
			var item FurnitureItem
			var createdAt string
			err := rows.Scan(&item.ID, &item.Name, &item.Level,
				&item.Price, &item.MaxLevel, &createdAt)
			if err != nil {
				continue
			}

			item.CreatedAt = createdAt
			furniture = append(furniture, item)
			totalValue += item.Price
			itemCount++
		}

		// 3. Возвращаем результат
		c.JSON(http.StatusOK, gin.H{
			"owner": gin.H{
				"id":       userID,
				"username": username,
			},
			"furniture": furniture,
			"stats": gin.H{
				"total_items": itemCount,
				"total_value": totalValue,
			},
		})
	}
}
