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

	// Create roles table
	createRolesTableSQL := `CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
	
	CREATE TRIGGER IF NOT EXISTS update_roles_timestamp 
	AFTER UPDATE ON roles
	BEGIN
		UPDATE roles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`

	_, err = DB.Exec(createRolesTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create permissions table
	createPermissionsTableSQL := `CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);
	
	-- Insert default permissions if they don't exist
	INSERT OR IGNORE INTO permissions (name, description) VALUES 
		('create_item', 'Ability to create new items'),
		('read_item', 'Ability to view items'),
		('update_item', 'Ability to edit existing items'),
		('delete_item', 'Ability to delete items'),
		('manage_roles', 'Ability to manage roles and permissions');`

	_, err = DB.Exec(createPermissionsTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create role_permissions table
	createRolePermissionsTableSQL := `CREATE TABLE IF NOT EXISTS role_permissions (
		role_id INTEGER NOT NULL,
		permission_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (role_id, permission_id),
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
		FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_role_permissions ON role_permissions(role_id, permission_id);`

	_, err = DB.Exec(createRolePermissionsTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create users table if it doesn't exist
	createUsersTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		role_id INTEGER,
		reset_token VARCHAR(255),
		reset_token_expires DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (role_id) REFERENCES roles(id)
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token);
	CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);

	CREATE TRIGGER IF NOT EXISTS update_users_timestamp 
	AFTER UPDATE ON users
	BEGIN
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;

	-- Create admin role if it doesn't exist
	INSERT OR IGNORE INTO roles (name, description) VALUES ('admin', 'Administrator with full access');
	
	-- Give admin role all permissions
	INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
	SELECT 
		(SELECT id FROM roles WHERE name = 'admin'),
		id
	FROM permissions;
	
	-- Create user role if it doesn't exist
	INSERT OR IGNORE INTO roles (name, description) VALUES ('user', 'Standard user with basic permissions');
	
	-- Give user role all basic CRUD permissions
	INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
	SELECT 
    	(SELECT id FROM roles WHERE name = 'user'),
    	id
	FROM permissions WHERE name IN ('read_item', 'create_item', 'update_item', 'delete_item');`

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
		UPDATE items SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`

	_, err = DB.Exec(createItemsTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}
