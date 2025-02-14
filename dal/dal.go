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

	// Create items table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        description TEXT
    );`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}
