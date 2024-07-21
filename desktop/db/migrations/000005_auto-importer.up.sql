CREATE TABLE IF NOT EXISTS transaction_importer_status
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS transaction_importer
(
    id INTEGER NOT NULL PRIMARY KEY,
    status_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    filesize INTEGER NOT NULL,
    checksum TEXT NOT NULL,
    date_created TEXT NOT NULL,
    date_imported TEXT NOT NULL,
    UNIQUE(checksum),
    FOREIGN KEY(status_id) REFERENCES transaction_importer_status(id)
);
