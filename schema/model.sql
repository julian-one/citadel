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
  source_type TEXT,
  source TEXT,
  prep_time INTEGER,
  cook_time INTEGER,
  serves INTEGER,
  cuisine TEXT,
  category TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS recipe_components (
  component_id TEXT PRIMARY KEY,
  recipe_id TEXT NOT NULL,
  name TEXT,
  position INTEGER NOT NULL DEFAULT 0,
  FOREIGN KEY (recipe_id) REFERENCES recipes (recipe_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ingredients (
  ingredient_id TEXT PRIMARY KEY,
  component_id TEXT NOT NULL,
  amount REAL NOT NULL,
  unit TEXT NOT NULL,
  item TEXT NOT NULL,
  FOREIGN KEY (component_id) REFERENCES recipe_components (component_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS instructions (
  instruction_id TEXT PRIMARY KEY,
  component_id TEXT NOT NULL,
  step_number INTEGER NOT NULL,
  instruction TEXT NOT NULL,
  FOREIGN KEY (component_id) REFERENCES recipe_components (component_id) ON DELETE CASCADE
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

CREATE TABLE IF NOT EXISTS recipe_reviews (
  review_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  recipe_id TEXT NOT NULL,
  rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
  difficulty INTEGER CHECK (difficulty >= 1 AND difficulty <= 5),
  duration INTEGER,
  notes TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (recipe_id) REFERENCES recipes (recipe_id) ON DELETE CASCADE
);

-- Enforce one review per user per recipe per calendar day.
CREATE UNIQUE INDEX IF NOT EXISTS idx_recipe_reviews_user_recipe_day
ON recipe_reviews (user_id, recipe_id, date(created_at));

CREATE TABLE IF NOT EXISTS pokemon (
  pokemon_id TEXT PRIMARY KEY,
  name text NOT NULL UNIQUE,
  height INTEGER NOT NULL,
  weight INTEGER NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS trading_sessions (
  session_id TEXT PRIMARY KEY,
  strategy TEXT NOT NULL,
  status TEXT NOT NULL,
  symbols TEXT NOT NULL,
  starting_capital REAL NOT NULL,
  parameters TEXT,
  started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  ended_at DATETIME
);

CREATE TABLE IF NOT EXISTS trading_orders (
  order_id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  client_order_id TEXT UNIQUE,
  symbol TEXT NOT NULL,
  side TEXT NOT NULL,
  type TEXT NOT NULL,
  qty REAL NOT NULL,
  filled_qty REAL DEFAULT 0,
  avg_price REAL,
  status TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES trading_sessions (session_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS trading_backtests (
  backtest_id TEXT PRIMARY KEY,
  strategy TEXT NOT NULL,
  symbols TEXT NOT NULL,
  start_date TEXT NOT NULL,
  end_date TEXT NOT NULL,
  starting_capital REAL NOT NULL,
  parameters TEXT,
  metrics TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
