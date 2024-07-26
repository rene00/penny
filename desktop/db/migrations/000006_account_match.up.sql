CREATE TABLE IF NOT EXISTS account_match
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    source_account_id INTEGER NOT NULL,
    destination_account_id INTEGER NOT NULL,
    UNIQUE(name, source_account_id, destination_account_id),
    FOREIGN KEY(source_account_id) REFERENCES account(id),
    FOREIGN KEY(destination_account_id) REFERENCES account(id)
);

CREATE TABLE IF NOT EXISTS account_match_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    account_match_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(account_match_id, name),
    FOREIGN KEY(account_match_id) REFERENCES account_match(id)
);

CREATE TABLE IF NOT EXISTS account_match_regex
(
    id INTEGER NOT NULL PRIMARY KEY,
    account_match_id INTEGER NOT NULL,
    regex TEXT NOT NULL,
    UNIQUE(account_match_id, regex),
    FOREIGN KEY(account_match_id) REFERENCES account_match(id)
);
