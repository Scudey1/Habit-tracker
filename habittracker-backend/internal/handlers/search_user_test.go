package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSearchUsersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createRequest := func(query string) *http.Request {
		req := httptest.NewRequest("GET", "/search", nil)
		if query != "" {
			q := req.URL.Query()
			q.Set("q", query)
			req.URL.RawQuery = q.Encode()
		}
		return req
	}

	t.Run("1. Успешный поиск пользователей", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(1, "TestUser1").
			AddRow(2, "TestUser2")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("test") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("test")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{
			"users": [
				{"id": 1, "username": "TestUser1"},
				{"id": 2, "username": "TestUser2"}
			],
			"count": 2
		}`, w.Body.String())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("2. Пустой поисковый запрос", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"users"`)
		assert.Contains(t, w.Body.String(), `"count"`)
		assert.Contains(t, w.Body.String(), `0`)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("3. Поиск с одним символом (слишком короткий)", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("a")

		handler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Минимум 2 символа для поиска")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("4. Поиск с пробелами", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(3, "Test User")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("test user") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("  test user  ")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Test User")
		assert.Contains(t, w.Body.String(), `"count":1`)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("5. Ошибка выполнения SQL запроса", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("error") + "%").
			WillReturnError(sql.ErrConnDone)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("error")

		handler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Ошибка поиска")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("6. Регистронезависимый поиск", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(4, "testuser").
			AddRow(5, "TestUser").
			AddRow(6, "TESTUSER")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("TEST") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("TEST")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "testuser")
		assert.Contains(t, w.Body.String(), "TestUser")
		assert.Contains(t, w.Body.String(), "TESTUSER")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("7. Поиск с спецсимволами", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(7, "user_123")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("user_") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("user_")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user_123")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("8. Пустой результат поиска", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"})

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("nonexistent") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("nonexistent")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"count":0`)
		assert.Contains(t, w.Body.String(), `"users"`)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("9. Ошибка сканирования строк", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow("неправильный_id", "TestUser")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("scanerror") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("scanerror")

		handler(c)

		if w.Code == http.StatusOK {
			assert.Contains(t, w.Body.String(), `"count":0`)
			assert.Contains(t, w.Body.String(), `"users"`)
		} else {
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("10. Поиск с русскими символами", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(8, "Иван").
			AddRow(9, "иван123")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("Иван") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("Иван")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Иван")
		assert.Contains(t, w.Body.String(), "иван123")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSearchUsersHandler_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createRequest := func(query string) *http.Request {
		req := httptest.NewRequest("GET", "/search", nil)
		if query != "" {
			q := req.URL.Query()
			q.Set("q", query)
			req.URL.RawQuery = q.Encode()
		}
		return req
	}

	t.Run("11. Без параметра q", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"users"`)
		assert.Contains(t, w.Body.String(), `"count":0`)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("12. Поиск с SQL инъекцией", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		injection := "test'; DROP TABLE users; --"

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower(injection) + "%").
			WillReturnRows(sqlmock.NewRows([]string{"id", "username"}))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest(injection)

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("13. Очень длинный поисковый запрос", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		longQuery := strings.Repeat("a", 1000)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(10, "testuser")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower(longQuery) + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest(longQuery)

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "testuser")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("14. Поиск с процентами и подчеркиваниями", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		query := "user%_name"

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(11, "user%_name123")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("user%_name") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest(query)

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user%_name123")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("15. Поиск с кодированными символами", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		query := "test+user"
		decodedQuery := "test user"

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(12, "test user")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower(decodedQuery) + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := httptest.NewRequest("GET", "/search?q="+url.QueryEscape(query), nil)
		c.Request = req

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test user")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("17. Полный HTTP запрос через роутер", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		router := gin.New()
		router.GET("/search", SearchUsersHandler(db))

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(1, "TestUser")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("test") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "TestUser")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("18. Запрос с пробелами через роутер", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		router := gin.New()
		router.GET("/search", SearchUsersHandler(db))

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(2, "John Doe")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("john doe") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/search?q=john+doe", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "John Doe")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSearchUsersHandler_Additional(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createRequest := func(query string) *http.Request {
		req := httptest.NewRequest("GET", "/search", nil)
		if query != "" {
			q := req.URL.Query()
			q.Set("q", query)
			req.URL.RawQuery = q.Encode()
		}
		return req
	}

	t.Run("19. Поиск только пробелами", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("   ")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"count":0`)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("20. Поиск с китайскими символами", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(13, "张三")

		// Для китайских символов ToLower не меняет строку
		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs("张三%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("张三")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "张三")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("21. Поиск с emoji", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(14, "user😀")

		// Для emoji ToLower не меняет строку
		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs("user😀%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("user😀")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user😀")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("22. Ограничение 20 результатов", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		rows := sqlmock.NewRows([]string{"id", "username"})
		for i := 1; i <= 20; i++ {
			rows.AddRow(i, fmt.Sprintf("user%d", i))
		}

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("user") + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("user")

		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		assert.NotNil(t, result["users"])
		assert.Equal(t, float64(20), result["count"])

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("23. Поиск с табуляцией и переносами строк", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		query := "test\tuser\nname"
		trimmedQuery := "test\tuser\nname"

		rows := sqlmock.NewRows([]string{"id", "username"}).
			AddRow(15, "test user name")

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower(trimmedQuery) + "%").
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest(query)

		handler(c)

		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("24. Ошибка подключения к БД", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		handler := SearchUsersHandler(db)

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("dbdown") + "%").
			WillReturnError(fmt.Errorf("connection refused"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("dbdown")

		handler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Ошибка поиска")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Бенчмарк тест
func BenchmarkSearchUsersHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)

	createRequest := func(query string) *http.Request {
		req := httptest.NewRequest("GET", "/search", nil)
		if query != "" {
			q := req.URL.Query()
			q.Set("q", query)
			req.URL.RawQuery = q.Encode()
		}
		return req
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, mock, err := sqlmock.New()
		if err != nil {
			b.Fatalf("Ошибка создания мока БД: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id", "username"})
		for j := 1; j <= 20; j++ {
			rows.AddRow(j, fmt.Sprintf("user%d", j))
		}

		mock.ExpectQuery(`
            SELECT 
                id,
                username
            FROM users 
            WHERE LOWER\(username\) LIKE LOWER\(\?\)
            ORDER BY username
            LIMIT 20
        `).WithArgs(strings.ToLower("user") + "%").
			WillReturnRows(rows)

		handler := SearchUsersHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = createRequest("user")

		handler(c)

		if w.Code != http.StatusOK {
			b.Fatalf("Ошибка: %d", w.Code)
		}
	}
}
