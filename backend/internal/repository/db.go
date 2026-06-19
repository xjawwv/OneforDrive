package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		getEnv("DB_USER", "rsuser"),
		getEnv("DB_PASSWORD", "rspass"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "routestorage"),
	)

	var err error
	for i := 0; i < 30; i++ {
		DB, err = sql.Open("mysql", dsn)
		if err == nil {
			err = DB.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("Waiting for MySQL... attempt %d/30", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	createTables()
	log.Println("Connected to MySQL")
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			quota_limit BIGINT DEFAULT 10737418240,
			quota_used BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS drive_accounts (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			email VARCHAR(255) NOT NULL,
			access_token TEXT,
			refresh_token TEXT,
			token_expiry TIMESTAMP NULL,
			capacity_total BIGINT DEFAULT 0,
			capacity_used BIGINT DEFAULT 0,
			route_storage_folder_id VARCHAR(255) NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			name VARCHAR(255) NOT NULL,
			mime_type VARCHAR(255) DEFAULT 'application/octet-stream',
			size_total BIGINT DEFAULT 0,
			status VARCHAR(50) DEFAULT 'active',
			upload_progress INT DEFAULT 100,
			parent_id BIGINT NULL,
			is_folder BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES files(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS file_chunks (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			file_id BIGINT NOT NULL,
			chunk_index INT NOT NULL,
			chunk_size BIGINT DEFAULT 0,
			drive_file_id VARCHAR(255),
			account_id BIGINT,
			checksum VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
			FOREIGN KEY (account_id) REFERENCES drive_accounts(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS shared_links (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			file_id BIGINT NOT NULL,
			token VARCHAR(64) UNIQUE NOT NULL,
			expires_at DATETIME NULL,
			created_by BIGINT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
			FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			description VARCHAR(255),
			is_system BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			` + "`key`" + ` VARCHAR(150) UNIQUE NOT NULL,
			description VARCHAR(255),
			category VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id BIGINT NOT NULL,
			permission_id BIGINT NOT NULL,
			PRIMARY KEY (role_id, permission_id),
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id BIGINT NOT NULL,
			role_id BIGINT NOT NULL,
			PRIMARY KEY (user_id, role_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS feature_routes (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			path VARCHAR(255) UNIQUE NOT NULL,
			icon VARCHAR(100) NOT NULL DEFAULT 'Circle',
			enabled BOOLEAN DEFAULT TRUE,
			description TEXT,
			category VARCHAR(100) NOT NULL DEFAULT 'general',
			display_order INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("Failed to create table: %v\nQuery: %s", err, q)
		}
	}

	DB.Exec("ALTER TABLE files ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at")
	DB.Exec("ALTER TABLE files ADD COLUMN upload_progress INT DEFAULT 100 AFTER status")
	DB.Exec("ALTER TABLE drive_accounts ADD COLUMN route_storage_folder_id VARCHAR(255) NULL AFTER capacity_used")

	seedRBAC()
	seedFeatureRoutes()
	migrateFeatureRoutes()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func seedRBAC() {
	DB.Exec("INSERT IGNORE INTO roles (name, description, is_system) VALUES ('owner', 'Full system access', TRUE)")
	DB.Exec("INSERT IGNORE INTO roles (name, description, is_system) VALUES ('admin', 'Manage users and all storage', TRUE)")
	DB.Exec("INSERT IGNORE INTO roles (name, description, is_system) VALUES ('member', 'Standard user, own storage only', TRUE)")

	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('users.manage', 'Create, delete, promote users', 'users')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('users.view_all', 'View all users', 'users')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('drive_accounts.manage_own', 'Connect/remove own Drive accounts', 'drive_accounts')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('drive_accounts.manage_all', 'Connect/remove any user Drive accounts', 'drive_accounts')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('files.upload', 'Upload files', 'files')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('files.delete_own', 'Delete own files', 'files')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('files.delete_all', 'Delete any user file', 'files')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('files.view_all', 'View all users files', 'files')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('storage.view_stats', 'View system-wide storage stats', 'storage')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('nav.explorer', 'Access Explorer page', 'navigation')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('nav.settings', 'Access Settings page', 'navigation')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('nav.admin', 'Access Admin section', 'navigation')")
	DB.Exec("INSERT IGNORE INTO permissions (`key`, description, category) VALUES ('nav.route_management', 'Access Route Management', 'navigation')")

	DB.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) SELECT (SELECT id FROM roles WHERE name = 'owner'), id FROM permissions")
	DB.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) SELECT (SELECT id FROM roles WHERE name = 'admin'), id FROM permissions WHERE `key` NOT IN ('nav.admin')")
	DB.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) SELECT (SELECT id FROM roles WHERE name = 'member'), id FROM permissions WHERE `key` IN ('drive_accounts.manage_own', 'files.upload', 'files.delete_own', 'nav.explorer', 'nav.settings')")

	DB.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) SELECT (SELECT id FROM roles WHERE name = 'owner'), id FROM permissions WHERE `key` LIKE 'nav.%'")
	DB.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) SELECT (SELECT id FROM roles WHERE name = 'admin'), id FROM permissions WHERE `key` IN ('nav.explorer', 'nav.settings')")

	log.Println("RBAC seed data inserted")
}

func seedFeatureRoutes() {
	routes := []struct {
		name, path, icon, desc, category string
		order                            int
	}{
		{"Explorer", "/explorer", "FolderOpen", "Browse and manage your files", "files", 1},
		{"Settings", "/settings", "Settings", "Manage your connected Google Drive accounts", "account", 2},
		{"Role Management", "/admin/roles", "ShieldCheck", "Manage roles and permissions", "admin", 3},
		{"User Management", "/admin/users", "Users", "View and manage user roles", "admin", 4},
		{"Route Management", "/admin/routes", "Map", "Enable or disable features", "admin", 5},
	}
	for _, r := range routes {
		DB.Exec(
			"INSERT IGNORE INTO feature_routes (name, path, icon, description, category, display_order) VALUES (?, ?, ?, ?, ?, ?)",
			r.name, r.path, r.icon, r.desc, r.category, r.order,
		)
	}
	log.Println("Feature routes seeded")
}

func migrateFeatureRoutes() {
	// Add exempt_role_ids column if it doesn't exist
	DB.Exec("ALTER TABLE feature_routes ADD COLUMN exempt_role_ids JSON NULL")

	// Ensure /admin/routes has all roles exempt (never lock yourself out)
	rows, err := DB.Query("SELECT id FROM roles")
	if err != nil {
		return
	}
	defer rows.Close()

	var allRoleIDs []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			allRoleIDs = append(allRoleIDs, id)
		}
	}

	if len(allRoleIDs) > 0 {
		roleIDsJSON := fmt.Sprintf("[%s]", int64SliceToJSON(allRoleIDs))
		DB.Exec("UPDATE feature_routes SET exempt_role_ids = ? WHERE path = '/admin/routes'", roleIDsJSON)
	}

	log.Println("Feature routes migration complete")
}

func int64SliceToJSON(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%d", id)
	}
	return result
}
