CREATE TABLE IF NOT EXISTS note
(
    id INTEGER NOT NULL PRIMARY KEY,
    note TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS note_transactions
(
    id INTEGER NOT NULL PRIMARY KEY,
    note_id INTEGER NOT NULL,
    transactions_id INTEGER NOT NULL,
    UNIQUE(note_id, transactions_id),
    FOREIGN KEY(note_id) REFERENCES note(id),
    FOREIGN KEY(transactions_id) REFERENCES transactions(id)
);
