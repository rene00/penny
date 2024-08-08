CREATE TABLE IF NOT EXISTS tag
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS tag_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    tag_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(tag_id, name),
    FOREIGN KEY(tag_id) REFERENCES tag(id)
);

CREATE TABLE IF NOT EXISTS tag_transactions
(
    id INTEGER NOT NULL PRIMARY KEY,
    tag_id INTEGER NOT NULL,
    transactions_id INTEGER NOT NULL,
    UNIQUE(tag_id, transactions_id),
    FOREIGN KEY(tag_id) REFERENCES tag(id),
    FOREIGN KEY(transactions_id) REFERENCES transactions(id)
);

CREATE TABLE IF NOT EXISTS tag_regex
(
    id INTEGER NOT NULL PRIMARY KEY,
    tag_id INTEGER NOT NULL,
    regex TEXT NOT NULL,
    UNIQUE(tag_id, regex),
    FOREIGN KEY(tag_id) REFERENCES tag(id)
);
