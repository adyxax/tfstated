CREATE TABLE schema_version (
  version INTEGER NOT NULL
) STRICT;

CREATE TABLE accounts (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  salt BLOB NOT NULL,
  password_hash BLOB NOT NULL,
  is_admin INTEGER NOT NULL DEFAULT FALSE,
  created INTEGER NOT NULL DEFAULT (unixepoch()),
  last_login INTEGER NOT NULL DEFAULT (unixepoch()),
  settings BLOB NOT NULL
) STRICT;
CREATE UNIQUE INDEX accounts_username on accounts(username);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  created INTEGER NOT NULL DEFAULT (unixepoch()),
  updated INTEGER NOT NULL DEFAULT (unixepoch()),
  data TEXT NOT NULL,
  FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE
) STRICT;

CREATE TABLE states (
  id TEXT PRIMARY KEY,
  path TEXT NOT NULL,
  lock TEXT,
  created INTEGER DEFAULT (unixepoch()),
  updated INTEGER DEFAULT (unixepoch())
) STRICT;
CREATE UNIQUE INDEX states_path on states(path);

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
