package main

import (
	"log"
	"os"

	"habittracker/internal/config"
	"habittracker/internal/database"
	"habittracker/internal/handlers"
	"habittracker/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println(" No .env file found, using defaults")
	}
	// 1. Загружаем конфигурацию
	cfg := config.Load()
	_ = cfg // Пока не используем, но оставляем для будущего

	// 2. Подключаемся к SQLite БД
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 3. Проверяем/создаем таблицы
	log.Println("Checking database tables...")

	// 4. Настраиваем роутер
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 5. Регистрируем маршруты
	router.POST("/api/register", handlers.Register(db))
	router.POST("/api/login", handlers.Login(db))

	// Тестовый маршрут
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"db":      "sqlite",
			"message": "Habit Tracker API is running",
		})
	})

	// Защищенные маршруты (требуют авторизации)
	auth := router.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		// Пример защищенного маршрута
		auth.GET("/profile", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			username := c.GetString("username")
			var coins int
			db.QueryRow("SELECT coins FROM users WHERE id = $1", userID).Scan(&coins)

			c.JSON(200, gin.H{
				"user_id":  userID,
				"username": username,
				"coins":    coins,
			})
		})
		// Задачи
		auth.GET("/search", handlers.SearchUsersHandler(db))
		auth.GET("/rooms/:user_id", handlers.GetUserRoomHandler(db))
		auth.GET("/tasks", handlers.GetTasks(db))
		auth.POST("/tasks/submit", handlers.SubmitTasks(db))
		auth.GET("/stats", handlers.GetUserStats(db))
		auth.GET("/tasks/catalog", handlers.GetTaskCatalog(db))
		auth.POST("/tasks/select", handlers.SelectTask(db))
		auth.GET("/furniture/catalog", handlers.GetFurnitureCatalog(db))
		auth.POST("/furniture/buy", handlers.BuyFurniture(db))
		auth.GET("/furniture/my", handlers.GetMyFurniture(db))

	}

	// 6. Запускаем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Println("Database: ./data/habittracker.db")
	log.Println(" Ready to accept connections!")
	log.Println("\nAvailable endpoints:")
	log.Println("  POST /api/register    - Регистрация")
	log.Println("  POST /api/login       - Вход")
	log.Println("  GET  /api/health      - Проверка работы")
	log.Println("  GET  /api/profile     - Профиль (требует токен)")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
