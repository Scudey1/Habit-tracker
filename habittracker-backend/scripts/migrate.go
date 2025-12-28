package main

import (
    "fmt"
    "log"
    "os"
	"path/filepath"
    
    "habittracker/internal/database"
)



func main() {
    // Подключаемся к БД
    db, err := database.Connect()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

	migrationFiles, err := filepath.Glob("migrations/*.sql")
    if err != nil {
        log.Fatal("Failed to list migration files:", err)
    }
    
 // Сортируем по имени (чтобы выполнять по порядку)
    for _, file := range migrationFiles {
        fmt.Printf("Running migration: %s\n", file)
        
        // Читаем файл миграции
        migrationSQL, err := os.ReadFile(file)
        if err != nil {
            log.Fatal("Failed to read migration file:", err)
        }
        
        // Выполняем миграцию
        _, err = db.Exec(string(migrationSQL))
        if err != nil {
            log.Fatal("Migration failed:", err)
        }
        
        fmt.Printf(" %s completed\n", file)
    }
    
    // Проверяем создание таблицы
    rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
    if err != nil {
        log.Fatal("Failed to check tables:", err)
    }
    defer rows.Close()
    
    fmt.Println("\nCreated tables:")
    for rows.Next() {
        var tableName string
        rows.Scan(&tableName)
        fmt.Println("  -", tableName)
    }
    
    fmt.Println("\nMigration completed successfully!")
    fmt.Println("Database file: ./data/habittracker.db")
}