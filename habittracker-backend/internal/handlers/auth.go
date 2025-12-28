package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"habittracker/internal/models"
	"habittracker/pkg/utils"

	"github.com/gin-gonic/gin"
)

//Login обрабатывает вход пользователя

func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некорректные данные",
			})
			return
		}

		// Ищем пользователя в базе
		var user models.User
		err := db.QueryRow(
			"SELECT id, username, password, coins FROM users WHERE username = ?",
			req.Username,
		).Scan(&user.ID, &user.Username, &user.Password, &user.Coins)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Неверное имя пользователя или пароль",
			})
			return
		}

		if err != nil {
			log.Printf("DB error (login): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка базы данных",
			})
			return
		}

		// Проверяем пароль
		if !user.CheckPassword(req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Неверное имя пользователя или пароль",
			})
			return
		}

		// Генерируем JWT токен
		token, err := utils.GenerateToken(user.ID, user.Username)
		if err != nil {
			log.Printf("Token generation error (login): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка создания токена",
			})
			return
		}

		// Возвращаем успешный ответ с токеном
		c.JSON(http.StatusOK, gin.H{
			"message": "Успешный вход",
			"token":   token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"coins":    user.Coins,
			},
		})
	}
}

// Register обрабатывает регистрацию пользователя
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Получаем данные из запроса
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Некорректные данные",
				"details": err.Error(),
			})
			return
		}

		// 2. Проверяем, не занят ли username
		var exists bool
		err := db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)",
			req.Username,
		).Scan(&exists)

		if err != nil {
			log.Printf("DB error (register, check exists): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка при проверке пользователя",
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Пользователь с таким именем уже существует",
			})
			return
		}

		// 3. Создаем нового пользователя
		user := models.User{
			Username: req.Username,
			Coins:    100, // Начальный баланс
		}

		// Хешируем пароль
		if err := user.HashPassword(req.Password); err != nil {
			log.Printf("Password hash error (register): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Ошибка при создании пользователя",
			})
			return
		}

		// 4. Сохраняем в БД (для SQLite другой синтаксис)
		result, err := db.Exec(
			`INSERT INTO users (username, password, coins) 
             VALUES (?, ?, ?)`,
			user.Username, user.Password, user.Coins,
		)

		if err != nil {
			log.Printf("DB error (register, insert): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось создать пользователя: " + err.Error(),
			})
			return
		}

		// Получаем ID вставленной записи
		userID, err := result.LastInsertId()
		if err != nil {
			log.Printf("DB error (register, lastInsertId): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось получить ID пользователя",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Пользователь успешно создан",
			"user_id":  userID,
			"coins":    user.Coins,
			"username": user.Username,
		})
	}
}
