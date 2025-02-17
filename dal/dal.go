package dal

import (
	"database/sql"
	"log"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create users table if it doesn't exist
	createUsersTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		reset_token VARCHAR(255),
		reset_token_expires DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token);

	CREATE TRIGGER IF NOT EXISTS update_users_timestamp 
	AFTER UPDATE ON users
	BEGIN
		UPDATE users 
		SET updated_at = CURRENT_TIMESTAMP 
		WHERE id = NEW.id;
	END;`

	_, err = DB.Exec(createUsersTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create items table if it doesn't exist
	createItemsTableSQL := `CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		CHECK (LENGTH(name) > 0)
	);

	CREATE INDEX IF NOT EXISTS idx_items_user_id ON items(user_id);
	CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);

	CREATE TRIGGER IF NOT EXISTS update_items_timestamp 
	AFTER UPDATE ON items
	BEGIN
		UPDATE items 
		SET updated_at = CURRENT_TIMESTAMP 
		WHERE id = NEW.id;
	END;`

	_, err = DB.Exec(createItemsTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}
