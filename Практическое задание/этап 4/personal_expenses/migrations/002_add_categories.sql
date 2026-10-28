CREATE TABLE IF NOT EXISTS categories (
	id	INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE
);

-- Добавление категории к трате
ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id);

-- Категории
INSERT INTO categories (name) SELECT 'Еда'
	WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = 'Еда');
INSERT INTO categories (name) SELECT 'Транспорт'
	WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = 'Транспорт');
INSERT INTO categories (name) SELECT 'Жильё'
	WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = 'Жильё');
INSERT INTO categories (name) SELECT 'Развлечения'
	WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = 'Развлечения');
INSERT INTO categories (name) SELECT 'Другое'
	WHERE NOT EXISTS (SELECT 1 FROM categories WHERE name = 'Другое');