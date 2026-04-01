CREATE TABLE IF NOT EXISTS users (
  user_id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  salt BLOB NOT NULL,
  role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
  session_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS posts (
  post_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  public BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS recipes (
  recipe_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  photo_url TEXT,
  source_url TEXT,
  cook_time INTEGER,
  serves INTEGER,
  cuisine TEXT,
  category TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ingredients (
  ingredient_id TEXT PRIMARY KEY,
  recipe_id TEXT NOT NULL,
  amount REAL NOT NULL,
  unit TEXT NOT NULL,
  item TEXT NOT NULL,
  FOREIGN KEY (recipe_id) REFERENCES recipes (recipe_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS instructions (
  instruction_id TEXT PRIMARY KEY,
  recipe_id TEXT NOT NULL,
  step_number INTEGER NOT NULL,
  instruction TEXT NOT NULL,
  FOREIGN KEY (recipe_id) REFERENCES recipes (recipe_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS recipe_bookmarks (
  bookmark_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  recipe_id TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (recipe_id) REFERENCES recipes (recipe_id) ON DELETE CASCADE,
  UNIQUE (user_id, recipe_id)
);

CREATE TABLE IF NOT EXISTS recipe_logs (
  log_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  recipe_id TEXT NOT NULL,
  notes TEXT,
  rating REAL CHECK (rating >= 1.0 AND rating <= 5.0),
  duration INTEGER,
  intensity INTEGER CHECK (intensity IN (1, 2, 3)),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (recipe_id) REFERENCES recipes (recipe_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pokemon (
  pokemon_id TEXT PRIMARY KEY,
  name text NOT NULL UNIQUE,
  height INTEGER NOT NULL,
  weight INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
