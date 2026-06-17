#!/usr/bin/env python3
"""
RouteStorage Database CLI
Usage: python db.py <command> [args]

Commands:
  status          Show database stats
  users           List all users
  roles           List all roles
  permissions     List all permissions
  user-roles      Show user-role assignments
  role-perms      Show role-permission assignments
  add-role <name> <desc>    Create a new role
  set-perm <role_id> <perm_ids...>   Set permissions for a role
  assign-role <user_id> <role_id>     Assign role to user
  remove-role <user_id> <role_id>     Remove role from user
  reset           Drop and recreate all tables
"""

import subprocess
import sys
import os

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = os.environ.get("DB_PORT", "3306")
DB_USER = os.environ.get("DB_USER", "rsuser")
DB_PASS = os.environ.get("DB_PASSWORD", "rspass")
DB_NAME = os.environ.get("DB_NAME", "routestorage")

def run_sql(sql):
    result = subprocess.run(
        ["docker", "compose", "exec", "-T", "mysql", "mysql",
         f"-u{DB_USER}", f"-p{DB_PASS}", DB_NAME, "-e", sql],
        capture_output=True, text=True, cwd=os.path.dirname(os.path.abspath(__file__))
    )
    if result.returncode != 0 and result.stderr:
        print(f"Error: {result.stderr.strip()}")
    return result.stdout.strip()

def cmd_status():
    print("=== Database Status ===")
    print(run_sql("SELECT COUNT(*) as users FROM users;"))
    print(run_sql("SELECT COUNT(*) as files FROM files WHERE is_folder = FALSE;"))
    print(run_sql("SELECT COUNT(*) as folders FROM files WHERE is_folder = TRUE;"))
    print(run_sql("SELECT COUNT(*) as drive_accounts FROM drive_accounts;"))
    print(run_sql("SELECT COUNT(*) as shared_links FROM shared_links;"))
    print(run_sql("SELECT COUNT(*) as roles FROM roles;"))
    print(run_sql("SELECT COUNT(*) as permissions FROM permissions;"))

def cmd_users():
    print(run_sql("SELECT id, email, name FROM users;"))

def cmd_roles():
    print(run_sql("SELECT id, name, description, is_system FROM roles;"))

def cmd_permissions():
    print(run_sql("SELECT id, `key`, description, category FROM permissions;"))

def cmd_user_roles():
    print(run_sql("""
        SELECT u.id, u.email, r.name as role_name
        FROM users u
        LEFT JOIN user_roles ur ON u.id = ur.user_id
        LEFT JOIN roles r ON ur.role_id = r.id
        ORDER BY u.id;
    """))

def cmd_role_perms():
    print(run_sql("""
        SELECT r.name as role_name, p.`key`, p.description
        FROM role_permissions rp
        JOIN roles r ON rp.role_id = r.id
        JOIN permissions p ON rp.permission_id = p.id
        ORDER BY r.name, p.`key`;
    """))

def cmd_add_role(args):
    if len(args) < 2:
        print("Usage: db.py add-role <name> <description>")
        return
    name, desc = args[0], " ".join(args[1:])
    print(run_sql(f"INSERT INTO roles (name, description) VALUES ('{name}', '{desc}');"))
    print(f"Role '{name}' created.")

def cmd_set_perm(args):
    if len(args) < 2:
        print("Usage: db.py set-perm <role_id> <perm_id1> [perm_id2] ...")
        return
    role_id = args[0]
    perm_ids = args[1:]
    sql = f"DELETE FROM role_permissions WHERE role_id = {role_id}; "
    for pid in perm_ids:
        sql += f"INSERT INTO role_permissions (role_id, permission_id) VALUES ({role_id}, {pid}); "
    print(run_sql(sql))
    print(f"Permissions set for role {role_id}.")

def cmd_assign_role(args):
    if len(args) < 2:
        print("Usage: db.py assign-role <user_id> <role_id>")
        return
    user_id, role_id = args[0], args[1]
    print(run_sql(f"INSERT IGNORE INTO user_roles (user_id, role_id) VALUES ({user_id}, {role_id});"))
    print(f"Role {role_id} assigned to user {user_id}.")

def cmd_remove_role(args):
    if len(args) < 2:
        print("Usage: db.py remove-role <user_id> <role_id>")
        return
    user_id, role_id = args[0], args[1]
    print(run_sql(f"DELETE FROM user_roles WHERE user_id = {user_id} AND role_id = {role_id};"))
    print(f"Role {role_id} removed from user {user_id}.")

def cmd_reset():
    confirm = input("This will DROP ALL TABLES. Type 'yes' to confirm: ")
    if confirm != "yes":
        print("Aborted.")
        return
    print("Dropping tables...")
    print(run_sql("SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS file_chunks, files, shared_links, drive_accounts, users, role_permissions, user_roles, permissions, roles; SET FOREIGN_KEY_CHECKS=1;"))
    print("All tables dropped. Restart the backend to recreate them.")
    print("Run: docker compose restart backend")

COMMANDS = {
    "status": cmd_status,
    "users": cmd_users,
    "roles": cmd_roles,
    "permissions": cmd_permissions,
    "user-roles": cmd_user_roles,
    "role-perms": cmd_role_perms,
    "add-role": cmd_add_role,
    "set-perm": cmd_set_perm,
    "assign-role": cmd_assign_role,
    "remove-role": cmd_remove_role,
    "reset": cmd_reset,
}

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(0)

    cmd = sys.argv[1]
    args = sys.argv[2:]

    if cmd in COMMANDS:
        if cmd in ("status", "users", "roles", "permissions", "user-roles", "role-perms"):
            COMMANDS[cmd]()
        else:
            COMMANDS[cmd](args)
    else:
        print(f"Unknown command: {cmd}")
        print(__doc__)
