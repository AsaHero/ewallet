CREATE TABLE IF NOT EXISTS categories(
    id integer,
    slug varchar(64) NOT NULL UNIQUE,
    position integer,
    name varchar(255) NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS categories_position_idx ON categories(position);

-- Categories INSERT query
INSERT INTO categories(id, slug, position, name)
VALUES
    (1, 'food-dining', 1, 'Food & Dining'),
(2, 'transport', 2, 'Transport'),
(3, 'groceries', 3, 'Groceries'),
(4, 'shopping', 4, 'Shopping'),
(5, 'entertainment', 5, 'Entertainment'),
(6, 'health-medical', 6, 'Health & Medical'),
(7, 'housing', 7, 'Housing'),
(8, 'utilities', 8, 'Utilities'),
(9, 'education', 9, 'Education'),
(10, 'personal-care', 10, 'Personal Care'),
(11, 'travel', 11, 'Travel'),
(12, 'gifts-donations', 12, 'Gifts & Donations'),
(13, 'insurance', 13, 'Insurance'),
(14, 'investments', 14, 'Investments'),
(15, 'salary', 15, 'Salary'),
(16, 'freelance', 16, 'Freelance'),
(17, 'business-income', 17, 'Business Income'),
(18, 'refunds', 18, 'Refunds'),
(19, 'fees-charges', 19, 'Fees & Charges'),
(20, 'subscriptions', 20, 'Subscriptions'),
(21, 'pets', 21, 'Pets'),
(22, 'sports-fitness', 22, 'Sports & Fitness'),
(23, 'bills', 23, 'Bills'),
(24, 'taxes', 24, 'Taxes'),
(25, 'other', 25, 'Other');

