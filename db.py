#!/usr/bin/env python3
"""
RouteStorage Database CLI
Usage: python db.py [command] [args]

Run with no arguments to launch the interactive menu (recommended).

Commands (still work directly, e.g. `python db.py status`):
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
  menu            Force the interactive menu

Requires: pip install rich
"""

import subprocess
import sys
import os

try:
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.prompt import Confirm
except ImportError:
    print("This tool needs the 'rich' package for colors and tables.")
    print("Install it with:  pip install rich")
    sys.exit(1)

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = os.environ.get("DB_PORT", "3306")
DB_USER = os.environ.get("DB_USER", "rsuser")
DB_PASS = os.environ.get("DB_PASSWORD", "rspass")
DB_NAME = os.environ.get("DB_NAME", "routestorage")

console = Console()


# ---------------------------------------------------------------------------
# Low-level helpers
# ---------------------------------------------------------------------------

def run_sql(sql):
    """Run SQL and return raw stdout. Errors are printed in red."""
    result = subprocess.run(
        ["docker", "compose", "exec", "-T", "mysql", "mysql",
         f"-u{DB_USER}", f"-p{DB_PASS}", DB_NAME, "-e", sql],
        capture_output=True, text=True, cwd=os.path.dirname(os.path.abspath(__file__))
    )
    if result.returncode != 0 and result.stderr:
        console.print(f"[bold red]Error:[/bold red] {result.stderr.strip()}")
    return result.stdout.strip()


def query_table(sql):
    """Run SQL and parse tab-separated output into a list of dicts."""
    result = subprocess.run(
        ["docker", "compose", "exec", "-T", "mysql", "mysql",
         f"-u{DB_USER}", f"-p{DB_PASS}", DB_NAME, "-e", sql],
        capture_output=True, text=True, cwd=os.path.dirname(os.path.abspath(__file__))
    )
    if result.returncode != 0:
        if result.stderr:
            console.print(f"[bold red]Error:[/bold red] {result.stderr.strip()}")
        return []
    lines = result.stdout.strip().split("\n")
    if not lines or not lines[0]:
        return []
    headers = lines[0].split("\t")
    rows = []
    for line in lines[1:]:
        values = line.split("\t")
        rows.append(dict(zip(headers, values)))
    return rows


def escape_sql(value):
    """Minimal escaping for string values interpolated into SQL."""
    return value.replace("\\", "\\\\").replace("'", "\\'")


def clear_screen():
    console.clear()


def print_table(title, rows, color="cyan"):
    """Render a list of dicts as a rich table. Shows a friendly message if empty."""
    if not rows:
        console.print(f"[yellow]No data found.[/yellow]")
        return
    table = Table(
        title=title,
        header_style=f"bold {color}",
        border_style=f"{color}",
        title_style="bold white",
        row_styles=["", "grey50"],
    )
    headers = list(rows[0].keys())
    for h in headers:
        table.add_column(h)
    for r in rows:
        values = []
        for h in headers:
            v = r.get(h, "")
            v = "-" if v == "NULL" else v
            values.append(str(v))
        table.add_row(*values)
    console.print(table)


def select_index(count, message="Select a number"):
    """Prompt for a 1-based index into a list of length `count`."""
    raw = console.input(f"\n[bold yellow]{message}:[/bold yellow] ").strip()
    if raw == "" or raw.lower() in ("q", "cancel", "back"):
        return None
    if not raw.isdigit():
        console.print("[red]Invalid input.[/red]")
        return None
    idx = int(raw)
    if not (1 <= idx <= count):
        console.print("[red]Number out of range.[/red]")
        return None
    return idx - 1


# ---------------------------------------------------------------------------
# Plain commands — now rendered as rich tables
# ---------------------------------------------------------------------------

def cmd_status():
    metrics = [
        ("Users", "SELECT COUNT(*) as c FROM users;"),
        ("Files", "SELECT COUNT(*) as c FROM files WHERE is_folder = FALSE;"),
        ("Folders", "SELECT COUNT(*) as c FROM files WHERE is_folder = TRUE;"),
        ("Drive Accounts", "SELECT COUNT(*) as c FROM drive_accounts;"),
        ("Shared Links", "SELECT COUNT(*) as c FROM shared_links;"),
        ("Roles", "SELECT COUNT(*) as c FROM roles;"),
        ("Permissions", "SELECT COUNT(*) as c FROM permissions;"),
    ]
    rows = []
    for label, sql in metrics:
        result = query_table(sql)
        value = result[0]["c"] if result else "?"
        rows.append({"Metric": label, "Count": value})
    print_table("Database Status", rows, color="green")


def cmd_users():
    rows = query_table("SELECT id, email, name FROM users ORDER BY id;")
    print_table("Users", rows, color="cyan")


def cmd_roles():
    rows = query_table("SELECT id, name, description, is_system FROM roles ORDER BY id;")
    print_table("Roles", rows, color="cyan")


def cmd_permissions():
    rows = query_table("SELECT id, `key`, description, category FROM permissions ORDER BY category, `key`;")
    print_table("Permissions", rows, color="cyan")


def cmd_user_roles():
    rows = query_table("""
        SELECT u.id, u.email, r.name as role_name
        FROM users u
        LEFT JOIN user_roles ur ON u.id = ur.user_id
        LEFT JOIN roles r ON ur.role_id = r.id
        ORDER BY u.id;
    """)
    print_table("User-Role Assignments", rows, color="magenta")


def cmd_role_perms():
    rows = query_table("""
        SELECT r.name as role_name, p.`key`, p.description
        FROM role_permissions rp
        JOIN roles r ON rp.role_id = r.id
        JOIN permissions p ON rp.permission_id = p.id
        ORDER BY r.name, p.`key`;
    """)
    print_table("Role-Permission Assignments", rows, color="magenta")


def cmd_add_role(args):
    if len(args) < 2:
        console.print("[yellow]Usage: db.py add-role <name> <description>[/yellow]")
        return
    name, desc = args[0], " ".join(args[1:])
    run_sql(
        f"INSERT INTO roles (name, description) VALUES "
        f"('{escape_sql(name)}', '{escape_sql(desc)}');"
    )
    console.print(f"[bold green]Role '{name}' created.[/bold green]")


def cmd_set_perm(args):
    if len(args) < 2:
        console.print("[yellow]Usage: db.py set-perm <role_id> <perm_id1> [perm_id2] ...[/yellow]")
        return
    role_id = args[0]
    perm_ids = args[1:]
    sql = f"DELETE FROM role_permissions WHERE role_id = {role_id}; "
    for pid in perm_ids:
        sql += f"INSERT INTO role_permissions (role_id, permission_id) VALUES ({role_id}, {pid}); "
    run_sql(sql)
    console.print(f"[bold green]Permissions set for role {role_id}.[/bold green]")


def cmd_assign_role(args):
    if len(args) < 2:
        console.print("[yellow]Usage: db.py assign-role <user_id> <role_id>[/yellow]")
        return
    user_id, role_id = args[0], args[1]
    run_sql(f"INSERT IGNORE INTO user_roles (user_id, role_id) VALUES ({user_id}, {role_id});")
    console.print(f"[bold green]Role {role_id} assigned to user {user_id}.[/bold green]")


def cmd_remove_role(args):
    if len(args) < 2:
        console.print("[yellow]Usage: db.py remove-role <user_id> <role_id>[/yellow]")
        return
    user_id, role_id = args[0], args[1]
    run_sql(f"DELETE FROM user_roles WHERE user_id = {user_id} AND role_id = {role_id};")
    console.print(f"[bold green]Role {role_id} removed from user {user_id}.[/bold green]")


def cmd_reset():
    confirmed = Confirm.ask("[bold red]This will DROP ALL TABLES. Continue?[/bold red]", default=False)
    if not confirmed:
        console.print("[yellow]Aborted.[/yellow]")
        return
    console.print("[red]Dropping tables...[/red]")
    run_sql("SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS file_chunks, files, shared_links, drive_accounts, users, role_permissions, user_roles, permissions, roles; SET FOREIGN_KEY_CHECKS=1;")
    console.print("[bold green]All tables dropped.[/bold green] Restart the backend to recreate them.")
    console.print("[dim]Run: docker compose restart backend[/dim]")


# ---------------------------------------------------------------------------
# Interactive wrappers — numbered selection instead of raw IDs
# ---------------------------------------------------------------------------

def interactive_add_role():
    console.print(Panel.fit("[bold]Add New Role[/bold]", border_style="green"))
    name = console.input("[bold yellow]Role name:[/bold yellow] ").strip()
    if not name:
        console.print("[red]Role name cannot be empty.[/red]")
        return
    desc = console.input("[bold yellow]Description:[/bold yellow] ").strip()
    cmd_add_role([name, desc])


def interactive_set_perm():
    roles = query_table("SELECT id, name FROM roles ORDER BY id;")
    if not roles:
        console.print("[yellow]No roles found.[/yellow]")
        return
    print_table("Step 1 — Choose a Role", roles, color="cyan")
    idx = select_index(len(roles), "Select a role")
    if idx is None:
        return
    role = roles[idx]

    perms = query_table("SELECT id, `key`, category FROM permissions ORDER BY category, `key`;")
    if not perms:
        console.print("[yellow]No permissions found.[/yellow]")
        return
    print_table(f"Step 2 — Permissions for '{role['name']}'", perms, color="magenta")
    raw = console.input("\n[bold yellow]Permission numbers, comma-separated (e.g. 1,3,4):[/bold yellow] ").strip()
    indices = []
    for part in raw.split(","):
        part = part.strip()
        if part.isdigit() and 1 <= int(part) <= len(perms):
            indices.append(int(part) - 1)
    if not indices:
        console.print("[red]No valid permissions selected. Aborted.[/red]")
        return
    perm_ids = [perms[i]["id"] for i in indices]
    cmd_set_perm([role["id"]] + perm_ids)


def interactive_assign_role():
    users = query_table("SELECT id, email FROM users ORDER BY id;")
    if not users:
        console.print("[yellow]No users found.[/yellow]")
        return
    print_table("Step 1 — Choose a User", users, color="cyan")
    uidx = select_index(len(users), "Select a user")
    if uidx is None:
        return
    user = users[uidx]

    roles = query_table("SELECT id, name FROM roles ORDER BY id;")
    if not roles:
        console.print("[yellow]No roles found.[/yellow]")
        return
    print_table(f"Step 2 — Choose a Role for '{user['email']}'", roles, color="magenta")
    ridx = select_index(len(roles), "Select a role")
    if ridx is None:
        return
    role = roles[ridx]

    cmd_assign_role([user["id"], role["id"]])


def interactive_remove_role():
    assignments = query_table("""
        SELECT u.id as user_id, u.email, r.id as role_id, r.name as role_name
        FROM user_roles ur
        JOIN users u ON ur.user_id = u.id
        JOIN roles r ON ur.role_id = r.id
        ORDER BY u.email;
    """)
    if not assignments:
        console.print("[yellow]No role assignments found.[/yellow]")
        return
    display_rows = [{"email": a["email"], "role": a["role_name"]} for a in assignments]
    print_table("Current Assignments", display_rows, color="magenta")
    idx = select_index(len(assignments), "Select an assignment to remove")
    if idx is None:
        return
    a = assignments[idx]
    cmd_remove_role([a["user_id"], a["role_id"]])


# ---------------------------------------------------------------------------
# Interactive menu
# ---------------------------------------------------------------------------

MENU_ITEMS = [
    ("Database status", cmd_status),
    ("List users", cmd_users),
    ("List roles", cmd_roles),
    ("List permissions", cmd_permissions),
    ("User-role assignments", cmd_user_roles),
    ("Role-permission assignments", cmd_role_perms),
    ("Add new role", interactive_add_role),
    ("Set permissions for a role", interactive_set_perm),
    ("Assign role to user", interactive_assign_role),
    ("Remove role from user", interactive_remove_role),
    ("Reset database (drop all tables)", cmd_reset),
]


def print_menu():
    console.print(Panel.fit(
        f"[bold white]RouteStorage DB CLI[/bold white]\n[dim]{DB_NAME}@{DB_HOST}:{DB_PORT}[/dim]",
        border_style="cyan",
    ))
    for i, (label, _) in enumerate(MENU_ITEMS, 1):
        console.print(f"  [bold cyan]{i:>2}.[/bold cyan] {label}")
    console.print(f"  [bold red] 0.[/bold red] Exit")


def main_menu():
    while True:
        clear_screen()
        print_menu()
        choice = console.input("\n[bold yellow]Select an option:[/bold yellow] ").strip()

        if choice in ("0", "q", "quit", "exit"):
            console.print("[dim]Bye.[/dim]")
            break
        if not choice.isdigit() or not (1 <= int(choice) <= len(MENU_ITEMS)):
            continue

        label, func = MENU_ITEMS[int(choice) - 1]
        clear_screen()
        console.print(Panel.fit(f"[bold white]{label}[/bold white]", border_style="green"))
        try:
            func()
        except Exception as e:
            console.print(f"[bold red]Error:[/bold red] {e}")
        console.input("\n[dim italic]Press Enter to continue...[/dim italic]")


# ---------------------------------------------------------------------------
# Entrypoint
# ---------------------------------------------------------------------------

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
    "menu": main_menu,
}

if __name__ == "__main__":
    if len(sys.argv) < 2:
        main_menu()
        sys.exit(0)

    cmd = sys.argv[1]
    args = sys.argv[2:]

    if cmd in ("-h", "--help", "help"):
        print(__doc__)
        sys.exit(0)

    if cmd in COMMANDS:
        if cmd in ("status", "users", "roles", "permissions", "user-roles", "role-perms", "reset", "menu"):
            COMMANDS[cmd]()
        else:
            COMMANDS[cmd](args)
    else:
        console.print(f"[red]Unknown command: {cmd}[/red]")
        print(__doc__)