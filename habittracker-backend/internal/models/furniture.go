package models

import "time"

type FurnitureTemplate struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Level     int    `json:"level"`
    Price     int    `json:"price"`
    ImageURL  string `json:"image_url"`
    MaxLevel  int    `json:"max_level"`
    CreatedAt time.Time `json:"created_at"`
}

type UserFurniture struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    TemplateID  int       `json:"template_id"`
    PurchaseDate time.Time `json:"purchase_date"`
    IsPlaced    bool      `json:"is_placed"`
    
    // Опционально
    Template *FurnitureTemplate `json:"template,omitempty"`
}

type UserUnlockedFurniture struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    TemplateID int      `json:"template_id"`
    UnlockedAt time.Time `json:"unlocked_at"`
}