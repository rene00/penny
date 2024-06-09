CREATE TABLE IF NOT EXISTS entity 
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    entity_type_id INTEGER NOT NULL,
    UNIQUE(name, entity_type_id)
);

CREATE TABLE IF NOT EXISTS entity_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    entity_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(entity_id, name),
    FOREIGN KEY(entity_id) REFERENCES entity(id)
);

CREATE TABLE IF NOT EXISTS entity_type
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INTEGER NOT NULL,
    UNIQUE(name, parent_id)
);

CREATE TABLE IF NOT EXISTS entity_type_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    entity_type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(entity_type_id, name),
    FOREIGN KEY(entity_type_id) REFERENCES entity_type(id)
);

CREATE TABLE IF NOT EXISTS account_type
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INTEGER NOT NULL,
    UNIQUE(name, parent_id)
);

CREATE TABLE IF NOT EXISTS account_type_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    account_type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(account_type_id, name),
    FOREIGN KEY(account_type_id) REFERENCES account_type(id)
);

CREATE TABLE IF NOT EXISTS account
(
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INTEGER NOT NULL,
    account_type_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    UNIQUE(name, parent_id, entity_id, account_type_id),
    FOREIGN KEY(entity_id) REFERENCES entity(id),
    FOREIGN KEY(account_type_id) REFERENCES account_type(id)
);

CREATE TABLE IF NOT EXISTS account_attributes
(
    id INTEGER NOT NULL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(account_id, name),
    FOREIGN KEY(account_id) REFERENCES account(id)
);

CREATE TABLE IF NOT EXISTS account_entity
(
    id INTEGER NOT NULL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    entity_id INTEGER NOT NULL,
    UNIQUE(account_id, entity_id),
    FOREIGN KEY(account_id) REFERENCES account(id),
    FOREIGN KEY(entity_id) REFERENCES entity(id)
);
