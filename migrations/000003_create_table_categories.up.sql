CREATE TABLE IF NOT EXISTS categories(
    id serial,
    user_id uuid,
    position integer,
    name_en varchar(255) NOT NULL,
    name_ru varchar(255) NOT NULL,
    name_uz varchar(255) NOT NULL,
    emoji varchar(64) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    PRIMARY KEY (id),
    CONSTRAINT categories_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS categories_position_idx ON categories(position ASC);

CREATE INDEX IF NOT EXISTS categories_user_id_idx ON categories(user_id);


CREATE TABLE IF NOT EXISTS subcategories(
    id serial,
    category_id integer NOT NULL,
    user_id uuid,
    position integer,
    name_en varchar(255) NOT NULL,
    name_ru varchar(255) NOT NULL,
    name_uz varchar(255) NOT NULL,
    emoji varchar(64) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone,
    PRIMARY KEY (id),
    CONSTRAINT subcategories_category_id_fk FOREIGN KEY (category_id) REFERENCES categories(id),
    CONSTRAINT subcategories_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS subcategories_position_idx ON subcategories(position ASC);

CREATE INDEX IF NOT EXISTS subcategories_user_id_idx ON subcategories(user_id);

INSERT INTO categories
(position, name_en, name_ru, name_uz, emoji)
VALUES
-- EXPENSES
(1,  'Food & Dining',         'Еда и питание',          'Ovqat va ovqatlanish', '🍽'),
(2,  'Transport',             'Транспорт',              'Transport',            '🚗'),
(3,  'Groceries',             'Продукты',               'Oziq-ovqat',           '🛒'),
(4,  'Shopping',              'Покупки',                'Xaridlar',             '🛍'),
(5,  'Entertainment',         'Развлечения',            'Ko‘ngilochar',         '🎉'),
(6,  'Health & Medical',      'Здоровье и медицина',    'Sog‘liq va tibbiyot',  '🏥'),
(7,  'Housing',               'Жильё',                  'Uy-joy',               '🏠'),
(8,  'Utilities',             'Коммунальные услуги',    'Kommunal xizmatlar',   '💡'),
(9,  'Education',             'Образование',            'Ta’lim',               '🎓'),
(10, 'Personal Care',         'Личная гигиена',         'Shaxsiy parvarish',    '🧴'),
(11, 'Travel',                'Путешествия',            'Sayohat',              '✈️'),
(12, 'Gifts & Donations',     'Подарки и пожертвования','Sovg‘alar va xayriya', '🎁'),
(13, 'Insurance',             'Страхование',            'Sug‘urta',             '🛡'),
(14, 'Investments',           'Инвестиции',             'Investitsiyalar',      '📈'),
(15, 'Salary',                'Зарплата',               'Ish haqi',             '💰'),
(16, 'Freelance',             'Фриланс',                'Frilans',              '🧑‍💻'),
(17, 'Business Income',       'Доход от бизнеса',       'Biznes daromadi',      '🏢'),
(18, 'Refunds',               'Возвраты',               'Qaytarimlar',          '↩️'),
(19, 'Fees & Charges',        'Комиссии и сборы',       'Komissiyalar',         '💸'),
(20, 'Subscriptions',         'Подписки',               'Obunalar',             '🔁'),
(21, 'Pets',                  'Домашние животные',      'Uy hayvonlari',        '🐾'),
(22, 'Sports & Fitness',      'Спорт и фитнес',         'Sport va fitnes',      '🏋️'),
(23, 'Taxes',                 'Налоги',                 'Soliqlar',             '🏛'),
(24, 'Communication',         'Связь',                  'Aloqa',                '📱'),
(25, 'Loans & Debts',         'Долги и займы',          'Qarzlar',              '💸'),
(26, 'Other',                 'Другое',                 'Boshqa',               '📦'),





INSERT INTO subcategories
(category_id, position, name_en, name_ru, name_uz, emoji)
VALUES
-- 🍽 Food & Dining (1)
(1, 1, 'Cafe', 'Кафе', 'Kafe', '☕'),
(1, 2, 'Restaurant', 'Ресторан', 'Restoran', '🍽'),
(1, 3, 'Fast Food', 'Фастфуд', 'Fast food', '🍔'),
(1, 4, 'Food Delivery', 'Доставка еды', 'Yetkazib berish', '🛵'),
(1, 5, 'Bar & Pub', 'Бар и паб', 'Bar va pab', '🍺'),
(1, 6, 'Bakery', 'Пекарня', 'Novvoyxona', '🥖'),
(1, 7, 'Street Food', 'Уличная еда', 'Ko‘cha ovqati', '🌮'),

-- 🚗 Transport (2)
(2, 1, 'Fuel', 'Бензин', 'Yoqilg‘i', '⛽'),
(2, 2, 'Taxi & Rideshare', 'Такси', 'Taksi', '🚕'),
(2, 3, 'Public Transport', 'Общественный транспорт', 'Jamoat transporti', '🚌'),
(2, 4, 'Car Maintenance', 'Обслуживание авто', 'Avto xizmat', '🛠'),
(2, 5, 'Parking & Tolls', 'Парковка и дороги', 'Avtoturargoh', '🅿️'),
(2, 6, 'Car Wash', 'Автомойка', 'Avto yuvish', '🧽'),
(2, 7, 'Car Insurance', 'Автостраховка', 'Avto sug‘urta', '🚗'),
(2, 8, 'Car Purchase/Lease', 'Покупка/аренда авто', 'Avto sotib olish', '🚙'),
(2, 9, 'Bicycle & Scooter', 'Велосипед и самокат', 'Velosiped va skuter', '🛴'),

-- 🛒 Groceries (3)
(3, 1, 'Supermarket', 'Супермаркет', 'Supermarket', '🛒'),
(3, 2, 'Market', 'Рынок', 'Bozor', '🥕'),
(3, 3, 'Convenience Store', 'Магазин у дома', 'Do‘kon', '🏪'),
(3, 4, 'Butcher', 'Мясная лавка', 'Go‘sht do‘koni', '🥩'),
(3, 5, 'Dairy Products', 'Молочные продукты', 'Sut mahsulotlari', '🥛'),
(3, 6, 'Fruits & Vegetables', 'Овощи и фрукты', 'Meva va sabzavot', '🍎'),

-- 🛍 Shopping (4)
(4, 1, 'Clothing', 'Одежда', 'Kiyim', '👕'),
(4, 2, 'Shoes', 'Обувь', 'Poyabzal', '👟'),
(4, 3, 'Electronics', 'Электроника', 'Elektronika', '📱'),
(4, 4, 'Household Goods', 'Товары для дома', 'Uy uchun', '🏠'),
(4, 5, 'Accessories', 'Аксессуары', 'Aksessuarlar', '👜'),
(4, 6, 'Jewelry', 'Ювелирные изделия', 'Zargarlik', '💍'),
(4, 7, 'Online Shopping', 'Интернет-покупки', 'Onlayn xarid', '🌐'),
(4, 8, 'Books & Stationery', 'Книги и канцтовары', 'Kitob va qirtosiya', '📚'),

-- 🎉 Entertainment (5)
(5, 1, 'Movies & Cinema', 'Кино', 'Kino', '🎬'),
(5, 2, 'Games', 'Игры', 'O‘yinlar', '🎮'),
(5, 3, 'Events & Concerts', 'Мероприятия', 'Tadbirlar', '🎟'),
(5, 4, 'Hobbies & Crafts', 'Хобби', 'Xobbi', '🎨'),
(5, 5, 'Music & Concerts', 'Музыка и концерты', 'Musiqa va konsertlar', '🎵'),
(5, 6, 'Sports Events', 'Спортивные события', 'Sport tadbirlari', '🏆'),
(5, 7, 'Theme Parks', 'Парки развлечений', 'O‘yin bog‘lari', '🎡'),

-- 🏥 Health & Medical (6)
(6, 1, 'Doctor Visit', 'Врач', 'Shifokor', '🩺'),
(6, 2, 'Pharmacy', 'Аптека', 'Dorixona', '💊'),
(6, 3, 'Dental', 'Стоматология', 'Stomatolog', '🦷'),
(6, 4, 'Medical Tests', 'Анализы', 'Tahlillar', '🧪'),
(6, 5, 'Hospital', 'Больница', 'Shifoxona', '🏥'),
(6, 6, 'Glasses & Lenses', 'Очки и линзы', 'Ko‘zoynak', '👓'),
(6, 7, 'Medical Devices', 'Медицинские приборы', 'Tibbiy qurilmalar', '🩹'),

-- 🏠 Housing (7)
(7, 1, 'Rent', 'Аренда', 'Ijara', '🏠'),
(7, 2, 'Mortgage', 'Ипотека', 'Ipoteka', '🏦'),
(7, 3, 'Hotel', 'Отель', 'Mehmonxona', '🏨'),
(7, 4, 'Home Repair', 'Ремонт', 'Ta‘mirlash', '🛠'),
(7, 5, 'Furniture', 'Мебель', 'Mebel', '🛋'),
(7, 6, 'Home Decor', 'Декор для дома', 'Uy bezagi', '🖼'),
(7, 7, 'Cleaning Services', 'Клининг', 'Tozalash xizmati', '🧹'),
(7, 8, 'Security', 'Охрана', 'Xavfsizlik', '🔒'),

-- 💡 Utilities (8)
(8, 1, 'Electricity', 'Электричество', 'Elektr', '💡'),
(8, 2, 'Gas', 'Газ', 'Gaz', '🔥'),
(8, 3, 'Water', 'Вода', 'Suv', '🚰'),
(8, 4, 'Heating', 'Отопление', 'Isitish', '🌡'),
(8, 5, 'Trash & Recycling', 'Вывоз мусора', 'Chiqindi', '🗑'),
(8, 6, 'HOA Fees', 'Коммунальные сборы', 'Kommunal to‘lovlar', '🏘'),

-- 🎓 Education (9)
(9, 1, 'Tuition', 'Обучение', 'O‘qish', '🎓'),
(9, 2, 'Courses & Training', 'Курсы', 'Kurslar', '📚'),
(9, 3, 'Books & Materials', 'Книги и материалы', 'Kitoblar', '📖'),
(9, 4, 'School Supplies', 'Школьные принадлежности', 'Maktab buyumlari', '✏️'),
(9, 5, 'Online Courses', 'Онлайн-курсы', 'Onlayn kurslar', '💻'),
(9, 6, 'Tutoring', 'Репетиторство', 'Repetitorlik', '👨‍🏫'),

-- 🧴 Personal Care (10)
(10, 1, 'Cosmetics & Skincare', 'Косметика', 'Kosmetika', '💄'),
(10, 2, 'Haircut & Salon', 'Парикмахерская', 'Sartaroshxona', '💇'),
(10, 3, 'Spa & Massage', 'Спа и массаж', 'Spa va massaj', '💆'),
(10, 4, 'Manicure & Pedicure', 'Маникюр и педикюр', 'Manikyur', '💅'),
(10, 5, 'Laundry & Dry Cleaning', 'Химчистка', 'Kimyoviy tozalash', '🧺'),
(10, 6, 'Personal Hygiene', 'Личная гигиена', 'Shaxsiy gigiena', '🧼'),

-- ✈️ Travel (11)
(11, 1, 'Flights', 'Авиабилеты', 'Aviabiletlar', '✈️'),
(11, 2, 'Accommodation', 'Проживание', 'Turar joy', '🏨'),
(11, 3, 'Transport', 'Транспорт', 'Transport', '🚌'),
(11, 4, 'Tours & Excursions', 'Туры и экскурсии', 'Sayohat', '🗺'),
(11, 5, 'Visa & Documents', 'Виза и документы', 'Viza', '📋'),
(11, 6, 'Travel Insurance', 'Страховка', 'Sug‘urta', '🛡'),
(11, 7, 'Luggage', 'Багаж', 'Yuк', '🧳'),

-- 🎁 Gifts & Donations (12)
(12, 1, 'Gifts', 'Подарки', 'Sovg‘alar', '🎁'),
(12, 2, 'Charity', 'Благотворительность', 'Xayriya', '❤️'),
(12, 3, 'Religious Donations', 'Религиозные пожертвования', 'Diniy xayriya', '🕌'),

-- 🛡 Insurance (13)
(13, 1, 'Health Insurance', 'Медицинская страховка', 'Tibbiy sug‘urta', '🏥'),
(13, 2, 'Life Insurance', 'Страхование жизни', 'Hayot sug‘urtasi', '🛡'),
(13, 3, 'Property Insurance', 'Страхование имущества', 'Mulk sug‘urtasi', '🏠'),
(13, 4, 'Other Insurance', 'Прочее страхование', 'Boshqa sug‘urta', '📋'),

-- 📈 Investments (14)
(14, 1, 'Stocks', 'Акции', 'Aksiyalar', '📈'),
(14, 2, 'Crypto', 'Криптовалюта', 'Kriptovalyuta', '🪙'),
(14, 3, 'Real Estate', 'Недвижимость', 'Ko‘chmas mulk', '🏢'),
(14, 4, 'Mutual Funds', 'Взаимные фонды', 'Investitsiya fondlari', '💼'),
(14, 5, 'Savings & Deposits', 'Сбережения и вклады', 'Omonatlar', '🏦'),

-- 💰 Salary (15)
(15, 1, 'Main Salary', 'Основная зарплата', 'Asosiy ish haqi', '💰'),
(15, 2, 'Bonus', 'Бонус', 'Bonus', '🎁'),
(15, 3, 'Overtime', 'Сверхурочные', 'Qo‘shimcha ish haqi', '⏰'),

-- 🧑‍💻 Freelance (16)
(16, 1, 'Project Payment', 'Оплата за проект', 'Loyiha to‘lovi', '🧑‍💻'),
(16, 2, 'Consultation', 'Консультации', 'Konsultatsiya', '💬'),
(16, 3, 'Royalties', 'Роялти', 'Royalti', '📝'),

-- 🏢 Business Income (17)
(17, 1, 'Sales Revenue', 'Доход от продаж', 'Savdo daromadi', '💵'),
(17, 2, 'Service Revenue', 'Доход от услуг', 'Xizmat daromadi', '🏢'),
(17, 3, 'Commission', 'Комиссионные', 'Komissiya', '💼'),

-- ↩️ Refunds (18)
(18, 1, 'Purchase Refund', 'Возврат покупки', 'Qaytarim', '↩️'),
(18, 2, 'Tax Refund', 'Возврат налога', 'Soliq qaytarish', '💵'),

-- 💸 Fees & Charges (19)
(19, 1, 'Bank Fees', 'Банковские комиссии', 'Bank komissiyasi', '🏦'),
(19, 2, 'Transaction Fees', 'Комиссия за перевод', 'O‘tkazma komissiyasi', '💸'),
(19, 3, 'Service Charges', 'Сервисные сборы', 'Xizmat to‘lovi', '📋'),
(19, 4, 'Late Fees', 'Штрафы за просрочку', 'Jarima', '⚠️'),

-- 🔁 Subscriptions (20)
(20, 1, 'Streaming Services', 'Стриминговые сервисы', 'Streaming xizmatlar', '📺'),
(20, 2, 'Software & Apps', 'ПО и приложения', 'Dasturlar', '💻'),
(20, 3, 'News & Media', 'Новости и медиа', 'Yangiliklar', '📰'),
(20, 4, 'Cloud Storage', 'Облачное хранилище', 'Bulutli xotira', '☁️'),
(20, 5, 'Fitness Apps', 'Фитнес-приложения', 'Fitnes ilovalar', '🏋️'),

-- 🐾 Pets (21)
(21, 1, 'Pet Food', 'Корм для питомцев', 'Uy hayvonlari ozuqasi', '🐾'),
(21, 2, 'Veterinary', 'Ветеринар', 'Veterinar', '🐶'),
(21, 3, 'Pet Supplies', 'Товары для питомцев', 'Hayvonlar uchun', '🦴'),
(21, 4, 'Pet Grooming', 'Груминг', 'Grooming', '✂️'),

-- 🏋️ Sports & Fitness (22)
(22, 1, 'Gym Membership', 'Спортзал', 'Sport zali', '🏋️'),
(22, 2, 'Sports Equipment', 'Спортинвентарь', 'Sport anjomlari', '⚽'),
(22, 3, 'Sports Classes', 'Спортивные занятия', 'Sport mashg‘ulotlari', '🤸'),
(22, 4, 'Personal Trainer', 'Персональный тренер', 'Shaxsiy murabbiy', '💪'),

-- 🏛 Taxes (23)
(23, 1, 'Income Tax', 'Подоходный налог', 'Daromad solig‘i', '💼'),
(23, 2, 'Property Tax', 'Налог на имущество', 'Mulk solig‘i', '🏠'),
(23, 3, 'Business Tax', 'Налог на бизнес', 'Biznes solig‘i', '🏢'),
(23, 4, 'VAT', 'НДС', 'QQS', '🧾'),

-- 📱 Communication (24)
(24, 1, 'Mobile Phone', 'Мобильная связь', 'Mobil aloqa', '📱'),
(24, 2, 'Internet', 'Интернет', 'Internet', '🌐'),
(24, 3, 'TV & Streaming', 'ТВ и стриминг', 'TV', '📺'),
(24, 4, 'Landline', 'Стационарный телефон', 'Statsionar telefon', '☎️'),

-- 📦 Other (25)
(25, 1, 'Miscellaneous', 'Разное', 'Turli xil', '📦'),
(25, 2, 'Cash Withdrawal', 'Снятие наличных', 'Naqd pul', '💵'),
(25, 3, 'Transfers', 'Переводы', 'O‘tkazmalar', '🔄'),
(25, 4, 'ATM Fees', 'Комиссия банкомата', 'Bankomat komissiyasi', '🏧');