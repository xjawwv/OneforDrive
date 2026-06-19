package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/routestorage/backend/internal/model"
	redispkg "github.com/routestorage/backend/pkg/redis"
)

func GetUserPermissions(userID int64) ([]string, error) {
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)
	if val, err := redispkg.Client.Get(ctx, cacheKey).Result(); err == nil {
		var perms []string
		json.Unmarshal([]byte(val), &perms)
		return perms, nil
	}

	rows, err := DB.Query(
		"SELECT DISTINCT p.`key` FROM permissions p JOIN role_permissions rp ON p.id = rp.permission_id JOIN user_roles ur ON rp.role_id = ur.role_id WHERE ur.user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err == nil {
			perms = append(perms, k)
		}
	}
	if perms == nil {
		perms = []string{}
	}

	data, _ := json.Marshal(perms)
	redispkg.Client.Set(ctx, cacheKey, string(data), 5*time.Minute)

	return perms, nil
}

func InvalidateUserPermissions(userID int64) {
	redispkg.Client.Del(ctx, fmt.Sprintf("user_permissions:%d", userID))
}

func GetUserRoles(userID int64) ([]model.Role, error) {
	rows, err := DB.Query(
		"SELECT r.id, r.name, r.description, r.is_system FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var r model.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem); err == nil {
			roles = append(roles, r)
		}
	}
	return roles, nil
}

func AssignRole(userID, roleID int64) error {
	_, err := DB.Exec("INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
	if err == nil {
		InvalidateUserPermissions(userID)
	}
	return err
}

func RemoveRole(userID, roleID int64) error {
	_, err := DB.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID)
	if err == nil {
		InvalidateUserPermissions(userID)
	}
	return err
}

func GetUserRoleIDs(userID int64) ([]int64, error) {
	rows, err := DB.Query("SELECT role_id FROM user_roles WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	if ids == nil {
		ids = []int64{}
	}
	return ids, nil
}

func ListRoles() ([]model.Role, error) {
	rows, err := DB.Query("SELECT id, name, description, is_system FROM roles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var r model.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem); err == nil {
			roles = append(roles, r)
		}
	}
	return roles, nil
}

func ListPermissions() ([]model.Permission, error) {
	rows, err := DB.Query("SELECT id, `key`, description, category FROM permissions ORDER BY category, `key`")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.Key, &p.Description, &p.Category); err == nil {
			perms = append(perms, p)
		}
	}
	return perms, nil
}

func CreateRole(name, description string) (int64, error) {
	result, err := DB.Exec("INSERT INTO roles (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetRolePermissions(roleID int64) ([]int64, error) {
	rows, err := DB.Query("SELECT permission_id FROM role_permissions WHERE role_id = ?", roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func SetRolePermissions(roleID int64, permissionIDs []int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID)
	for _, pid := range permissionIDs {
		tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, pid)
	}

	return tx.Commit()
}

var ctx = context.Background()
