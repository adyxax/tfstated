CREATE TABLE schema_version (
  version INTEGER NOT NULL
) STRICT;

CREATE TABLE accounts (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  salt BLOB,
  password_hash BLOB,
  is_admin INTEGER NOT NULL,
  created INTEGER NOT NULL DEFAULT (unixepoch()),
  last_login INTEGER NOT NULL DEFAULT (unixepoch()),
  settings BLOB NOT NULL,
  password_reset TEXT
) STRICT;
CREATE UNIQUE INDEX accounts_username ON accounts(username);

CREATE TABLE sessions (
  id BLOB PRIMARY KEY,
  created INTEGER NOT NULL DEFAULT (unixepoch()),
  updated INTEGER NOT NULL DEFAULT (unixepoch()),
  data BLOB NOT NULL
) STRICT;
CREATE INDEX sessions_data_account_id ON sessions(data->'account'->>'id');

CREATE TABLE states (
  id TEXT PRIMARY KEY,
  path TEXT NOT NULL,
  lock TEXT,
  created INTEGER DEFAULT (unixepoch()),
  updated INTEGER DEFAULT (unixepoch())
) STRICT;
CREATE UNIQUE INDEX states_path ON states(path);

CREATE TABLE versions (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  state_id TEXT,
  data BLOB,
  lock TEXT,
  created INTEGER DEFAULT (unixepoch()),
  FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE
  FOREIGN KEY(state_id) REFERENCES states(id) ON DELETE CASCADE
) STRICT;
