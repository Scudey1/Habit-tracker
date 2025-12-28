-- Шаблоны мебели (каталог с уровнями)
CREATE TABLE IF NOT EXISTS furniture_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL, -- Название: "Телевизор", "Стул" и т.д.
    level INTEGER NOT NULL DEFAULT 1, -- Уровень предмета (1, 2, 3)
    price INTEGER NOT NULL, -- Цена в монетах
    image_url TEXT, -- URL изображения
    max_level INTEGER DEFAULT 3, -- Максимальный уровень (когда исчезает)
    next_level_template_id INTEGER, -- ID следующего уровня (если есть)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (next_level_template_id) REFERENCES furniture_templates(id)
);

-- Купленная мебель пользователя
CREATE TABLE IF NOT EXISTS user_furniture (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL, -- Какой предмет куплен
    purchase_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_placed BOOLEAN DEFAULT 0, -- Размещен ли в комнате
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES furniture_templates(id)
);

-- Уровни разблокировки предметов для пользователя
CREATE TABLE IF NOT EXISTS user_unlocked_furniture (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL, -- Какой предмет разблокирован
    unlocked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES furniture_templates(id),
    UNIQUE(user_id, template_id)
);