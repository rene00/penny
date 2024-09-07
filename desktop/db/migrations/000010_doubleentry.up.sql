PRAGMA foreign_keys=off;

ALTER TABLE splits RENAME TO splits_1;

CREATE TABLE IF NOT EXISTS splits
(
    id INTEGER NOT NULL PRIMARY KEY,
    transactions_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    value_num BIGINT NOT NULL,
    value_denom BIGINT NOT NULL,
    FOREIGN KEY(transactions_id) REFERENCES transactions(id),
    FOREIGN KEY(account_id) REFERENCES account(id)
);

INSERT INTO splits SELECT * FROM splits_1;

DROP TABLE splits_1;

PRAGMA foreign_keys=on;
