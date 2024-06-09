CREATE TABLE IF NOT EXISTS transactions_hash_type (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS transactions_hash (
    id INTEGER NOT NULL PRIMARY KEY,
    transactions_id INTEGER NOT NULL,
    transactions_hash_type_id INTEGER NOT NULL,
    hash TEXT NOT NULL,
    UNIQUE(transactions_id, transactions_hash_type_id, hash),
    FOREIGN KEY(transactions_id) REFERENCES transactions(id),
    FOREIGN KEY(transactions_hash_type_id) REFERENCES transactions_hash_type(id)
);
