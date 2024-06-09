CREATE TABLE IF NOT EXISTS currency
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions
(
    id INTEGER NOT NULL PRIMARY KEY,
    currency_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    memo TEXT NOT NULL,
    date TEXT NOT NULL,
    FOREIGN KEY(currency_id) REFERENCES currency(id),
    FOREIGN KEY(entity_id) REFERENCES entity(id)
);

CREATE TABLE IF NOT EXISTS transactions_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    transactions_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(transactions_id, name),
    FOREIGN KEY(transactions_id) REFERENCES transactions(id)
);

CREATE TABLE IF NOT EXISTS splits
(
    id INTEGER NOT NULL PRIMARY KEY,
    transactions_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    value_num BIGINT NOT NULL,
    value_denom BIGINT NOT NULL,
    UNIQUE(transactions_id, account_id),
    FOREIGN KEY(transactions_id) REFERENCES transactions(id),
    FOREIGN KEY(account_id) REFERENCES account(id)
);

CREATE TABLE IF NOT EXISTS splits_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    splits_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(splits_id, name),
    FOREIGN KEY(splits_id) REFERENCES splits(id)
);
