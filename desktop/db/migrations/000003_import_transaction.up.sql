CREATE TABLE IF NOT EXISTS currency_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    currency_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(currency_id, name),
    FOREIGN KEY(currency_id) REFERENCES currency(id)
);
