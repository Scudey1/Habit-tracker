package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetFurnitureCatalog возвращает доступную для покупки мебель
func GetFurnitureCatalog(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		// Получаем каталог:
		// 1. Все предметы уровня 1, которые НЕ куплены
		// 2. ИЛИ предметы уровня 2, если куплен уровень 1 того же имени
		rows, err := db.Query(`
            WITH user_furniture_items AS (
                SELECT 
                    ft.name,
                    ft.level,
                    ft.max_level
                FROM user_furniture uf
                JOIN furniture_templates ft ON uf.template_id = ft.id
                WHERE uf.user_id = ?
            )
            SELECT 
                ft.id,
                ft.name,
                ft.level,
                ft.price,
                ft.max_level,
                ft.created_at,
                CASE 
                    WHEN ft.level = 1 THEN 1
                    WHEN ft.level = 2 AND EXISTS (
                        SELECT 1 FROM user_furniture_items ufi
                        WHERE ufi.name = ft.name AND ufi.level = 1
                    ) THEN 1
                    ELSE 0
                END as can_buy
            FROM furniture_templates ft
            WHERE ft.id NOT IN (
                SELECT uf.template_id 
                FROM user_furniture uf 
                WHERE uf.user_id = ?
            )
            AND (
                -- Уровень 1: показываем если не куплен и нет уровня 2
                (
                    ft.level = 1
                    AND NOT EXISTS (
                        SELECT 1 FROM user_furniture_items ufi
                        WHERE ufi.name = ft.name AND ufi.level = 2
                    )
                )
                -- ИЛИ уровень 2: показываем если есть уровень 1
                OR (
                    ft.level = 2 
                    AND EXISTS (
                        SELECT 1 FROM user_furniture_items ufi
                        WHERE ufi.name = ft.name AND ufi.level = 1
                    )
                )
            )
            ORDER BY ft.price
        `, userID, userID, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения каталога мебели: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		type CatalogItem struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Level     int    `json:"level"`
			Price     int    `json:"price"`
			MaxLevel  int    `json:"max_level"`
			CreatedAt string `json:"created_at"`
			CanBuy    bool   `json:"can_buy"` // true - можно купить сейчас
		}

		var catalog []CatalogItem
		for rows.Next() {
			var item CatalogItem
			var createdAt string
			var canBuy int
			rows.Scan(&item.ID, &item.Name, &item.Level, &item.Price,
				&item.MaxLevel, &createdAt, &canBuy)
			item.CreatedAt = createdAt
			item.CanBuy = canBuy == 1
			catalog = append(catalog, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"catalog": catalog,
			"total":   len(catalog),
		})
	}
}

// BuyFurniture покупка мебели (работает с уровнями 1 и 2)
func BuyFurniture(db *sql.DB) gin.HandlerFunc {
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

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка транзакции",
			})
			return
		}

		// 1. Получаем информацию о предмете
		var name string
		var level, price, maxLevel int
		err = tx.QueryRow(`
            SELECT name, level, price, max_level 
            FROM furniture_templates 
            WHERE id = ?
        `, req.TemplateID).Scan(&name, &level, &price, &maxLevel)

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Предмет не найден",
			})
			return
		}

		// 2. Проверяем баланс
		var userCoins int
		tx.QueryRow("SELECT coins FROM users WHERE id = ?", userID).Scan(&userCoins)
		if userCoins < price {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Недостаточно монет",
				"balance":  userCoins,
				"required": price,
			})
			return
		}

		// 3. Проверяем можно ли купить этот уровень
		if level == 1 {
			// Для уровня 1: нельзя если уже есть любой уровень с таким именем
			var alreadyPurchased bool
			tx.QueryRow(`
                SELECT COUNT(*) > 0 
                FROM user_furniture uf
                JOIN furniture_templates ft ON uf.template_id = ft.id
                WHERE uf.user_id = ? AND ft.name = ?
            `, userID, name).Scan(&alreadyPurchased)

			if alreadyPurchased {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Вы уже купили этот предмет",
					"hint":  "Можно купить только один предмет с данным названием",
				})
				return
			}
		} else if level == 2 {
			// Для уровня 2: можно купить ТОЛЬКО если есть уровень 1
			var hasLevel1 bool
			tx.QueryRow(`
                SELECT COUNT(*) > 0 
                FROM user_furniture uf
                JOIN furniture_templates ft ON uf.template_id = ft.id
                WHERE uf.user_id = ? AND ft.name = ? AND ft.level = 1
            `, userID, name).Scan(&hasLevel1)

			if !hasLevel1 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Нельзя купить уровень 2 без уровня 1",
				})
				return
			}

			// Проверяем что уровень 2 еще не куплен
			var hasLevel2 bool
			tx.QueryRow(`
                SELECT COUNT(*) > 0 
                FROM user_furniture uf
                JOIN furniture_templates ft ON uf.template_id = ft.id
                WHERE uf.user_id = ? AND ft.name = ? AND ft.level = 2
            `, userID, name).Scan(&hasLevel2)

			if hasLevel2 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Уровень 2 уже куплен",
				})
				return
			}
		} else {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректный уровень предмета",
			})
			return
		}

		// 4. Если покупаем уровень 2, находим и удаляем уровень 1
		var level1ID int
		if level == 2 {
			// Находим ID уровня 1
			err = tx.QueryRow(`
                SELECT ft.id 
                FROM user_furniture uf
                JOIN furniture_templates ft ON uf.template_id = ft.id
                WHERE uf.user_id = ? AND ft.name = ? AND ft.level = 1
            `, userID, name).Scan(&level1ID)

			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Не найден уровень 1 для замены",
				})
				return
			}

			// Удаляем уровень 1
			_, err = tx.Exec(`
                DELETE FROM user_furniture 
                WHERE user_id = ? AND template_id = ?
            `, userID, level1ID)

			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка удаления уровня 1",
				})
				return
			}
		}

		// 5. Списываем монеты
		_, err = tx.Exec("UPDATE users SET coins = coins - ? WHERE id = ?", price, userID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка списания монет",
			})
			return
		}

		// 6. Добавляем мебель пользователю
		_, err = tx.Exec(`
            INSERT INTO user_furniture (user_id, template_id) 
            VALUES (?, ?)
        `, userID, req.TemplateID)

		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка покупки",
			})
			return
		}

		// 7. Коммитим
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка завершения покупки",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Покупка успешна",
			"item": gin.H{
				"id":    req.TemplateID,
				"name":  name,
				"level": level,
				"price": price,
			},
			"spent_coins": price,
			"new_balance": userCoins - price,
		})
	}
}

// GetMyFurniture возвращает купленную мебель пользователя
// Показывает только максимальный уровень для каждого предмета
func GetMyFurniture(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		// Получаем максимальный купленный уровень для каждого предмета
		rows, err := db.Query(`
            SELECT 
                uf.id,
                uf.template_id,
                ft.name,
                ft.level,
                ft.price,
                ft.max_level,
                ft.created_at
            FROM user_furniture uf
            JOIN furniture_templates ft ON uf.template_id = ft.id
            WHERE uf.user_id = ?
            ORDER BY ft.name
        `, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения мебели: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		type FurnitureItem struct {
			ID         int    `json:"id"`
			TemplateID int    `json:"template_id"`
			Name       string `json:"name"`
			Level      int    `json:"level"`
			Price      int    `json:"price"`
			MaxLevel   int    `json:"max_level"`
			CreatedAt  string `json:"created_at"`
		}

		var furniture []FurnitureItem
		for rows.Next() {
			var item FurnitureItem
			var createdAt string
			rows.Scan(&item.ID, &item.TemplateID, &item.Name,
				&item.Level, &item.Price, &item.MaxLevel, &createdAt)
			item.CreatedAt = createdAt
			furniture = append(furniture, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"furniture": furniture,
			"total":     len(furniture),
		})
	}
}

// GetAvailableUpgrades возвращает доступные улучшения для купленной мебели
func GetAvailableUpgrades(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")

		// Находим купленную мебель уровня 1, у которой есть уровень 2
		rows, err := db.Query(`
            SELECT 
                ft1.id as level1_id,
                ft1.name,
                ft1.level as current_level,
                ft2.id as level2_id,
                ft2.price as upgrade_price,
                ft1.max_level
            FROM user_furniture uf
            JOIN furniture_templates ft1 ON uf.template_id = ft1.id
            LEFT JOIN furniture_templates ft2 ON ft1.name = ft2.name AND ft2.level = 2
            WHERE uf.user_id = ? 
              AND ft1.level = 1 
              AND ft2.id IS NOT NULL
              AND NOT EXISTS (
                  -- Проверяем что уровень 2 еще не куплен
                  SELECT 1 FROM user_furniture uf2
                  JOIN furniture_templates ft3 ON uf2.template_id = ft3.id
                  WHERE uf2.user_id = uf.user_id 
                    AND ft3.name = ft1.name 
                    AND ft3.level = 2
              )
        `, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка получения улучшений: " + err.Error(),
			})
			return
		}
		defer rows.Close()

		type UpgradeItem struct {
			Level1ID     int    `json:"level1_id"`
			Name         string `json:"name"`
			CurrentLevel int    `json:"current_level"`
			Level2ID     int    `json:"level2_id"`
			UpgradePrice int    `json:"upgrade_price"`
			MaxLevel     int    `json:"max_level"`
		}

		var upgrades []UpgradeItem
		for rows.Next() {
			var item UpgradeItem
			rows.Scan(&item.Level1ID, &item.Name, &item.CurrentLevel,
				&item.Level2ID, &item.UpgradePrice, &item.MaxLevel)
			upgrades = append(upgrades, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"upgrades": upgrades,
			"total":    len(upgrades),
		})
	}
}
