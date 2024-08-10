CREATE TABLE IF NOT EXISTS attachment
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    sz INTEGER NOT NULL,
    data BLOB NOT NULL,
    created_at INTEGER NOT NULL,
    UNIQUE(name, data)
);

CREATE TABLE IF NOT EXISTS attachment_transactions
(
    id INTEGER NOT NULL PRIMARY KEY,
    attachment_id INTEGER NOT NULL,
    transactions_id INTEGER NOT NULL,
    UNIQUE(attachment_id, transactions_id),
    FOREIGN KEY(attachment_id) REFERENCES attachment(id),
    FOREIGN KEY(transactions_id) REFERENCES transactions(id)
);
