CREATE TABLE IF NOT EXISTS task_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    base_reward INTEGER DEFAULT 1, -- Базовая награда в монетах
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Задачи пользователя (выбранные из каталога)
CREATE TABLE IF NOT EXISTS user_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL, -- Ссылка на шаблон
    is_active BOOLEAN DEFAULT 1, -- Активна ли задача
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES task_templates(id),
    UNIQUE(user_id, template_id) -- Чтобы не дублировать задачи
);

-- Прогресс по задачам
CREATE TABLE IF NOT EXISTS task_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_task_id INTEGER NOT NULL, -- Ссылка на user_tasks
    date DATE NOT NULL, -- Дата выполнения
    completed BOOLEAN DEFAULT 0, -- Выполнена ли в этот день
    streak_days INTEGER DEFAULT 0, -- Текущая серия дней
    total_reward INTEGER DEFAULT 0, -- Общая награда за этот день
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_task_id) REFERENCES user_tasks(id) ON DELETE CASCADE,
    UNIQUE(user_task_id, date) -- Одна запись на день
);