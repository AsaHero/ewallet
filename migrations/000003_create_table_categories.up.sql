CREATE TABLE IF NOT EXISTS categories(
    id integer,
    slug varchar(64) NOT NULL UNIQUE,
    position integer,
    name varchar(255) NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS categories_position_idx ON categories(position);

INSERT INTO categories(id, slug, position, name)
VALUES
    (1, 'food-dining', 1, 'Еда и питание'),
    (2, 'transport', 2, 'Транспорт'),
    (3, 'groceries', 3, 'Продукты'),
    (4, 'shopping', 4, 'Покупки'),
    (5, 'entertainment', 5, 'Развлечения'),
    (6, 'health-medical', 6, 'Здоровье и медицина'),
    (7, 'housing', 7, 'Жильё'),
    (8, 'utilities', 8, 'Коммунальные услуги'),
    (9, 'education', 9, 'Образование'),
    (10, 'personal-care', 10, 'Личная гигиена'),
    (11, 'travel', 11, 'Путешествия'),
    (12, 'gifts-donations', 12, 'Подарки и пожертвования'),
    (13, 'insurance', 13, 'Страхование'),
    (14, 'investments', 14, 'Инвестиции'),
    (15, 'salary', 15, 'Зарплата'),
    (16, 'freelance', 16, 'Фриланс'),
    (17, 'business-income', 17, 'Доход от бизнеса'),
    (18, 'refunds', 18, 'Возвраты'),
    (19, 'fees-charges', 19, 'Комиссии и сборы'),
    (20, 'subscriptions', 20, 'Подписки'),
    (21, 'pets', 21, 'Домашние животные'),
    (22, 'sports-fitness', 22, 'Спорт и фитнес'),
    (23, 'bills', 23, 'Счета'),
    (24, 'taxes', 24, 'Налоги'),
    (25, 'other', 25, 'Другое');
