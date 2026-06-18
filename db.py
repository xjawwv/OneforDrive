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
except ImportError:
    print("This tool needs the 'rich' package for colors and tables.")
    print("Install it with:  pip install rich")
    sys.exit(1)

DB_HOST = os.environ.get("DB_HOST", "localhost")
DB_PORT = os.environ.get("DB_PORT", "3306")
DB_USER = os.environ.get("DB_USER", "rsuser")
DB_PASS = os.environ.get("DB_PASSWORD", "rspass")
DB_NAME = os.environ.get("DB_NAME", "routestorage")

REDIS_HOST = os.environ.get("REDIS_HOST", "localhost")
REDIS_PORT = os.environ.get("REDIS_PORT", "6379")

console = Console()


# ---------------------------------------------------------------------------
# Keyboard input
# ---------------------------------------------------------------------------

def read_key():
    """Read a single keypress and return a normalized action string."""
    if sys.platform == "win32":
        import msvcrt
        key = msvcrt.getwch()
        if key in ('\x00', '\xe0'):
            key2 = msvcrt.getwch()
            return {"H": "up", "P": "down", "K": "left", "M": "right"}.get(key2, "")
        if key in ('\r', '\n'):
            return "enter"
        if key == ' ':
            return "space"
        if key == '\x1b':
            return "escape"
        if key == '\x03':
            return "ctrl_c"
        return key
    else:
        import tty, termios
        fd = sys.stdin.fileno()
        old = termios.tcgetattr(fd)
        try:
            tty.setraw(fd)
            ch = sys.stdin.read(1)
            if ch == '\x1b':
                ch2 = sys.stdin.read(1)
                if ch2 == '[':
                    ch3 = sys.stdin.read(1)
                    return {"A": "up", "B": "down", "C": "right", "D": "left"}.get(ch3, "")
                return "escape"
            if ch in ('\r', '\n', ' '):
                return "enter"
            if ch == '\x03':
                return "ctrl_c"
            return ch
        finally:
            termios.tcsetattr(fd, termios.TCSADRAIN, old)


def clear_screen():
    os.system('cls' if sys.platform == 'win32' else 'clear')


# ---------------------------------------------------------------------------
# Low-level helpers
# ---------------------------------------------------------------------------

def run_sql(sql):
    result = subprocess.run(
        ["docker", "compose", "exec", "-T", "mysql", "mysql",
         f"-u{DB_USER}", f"-p{DB_PASS}", DB_NAME, "-e", sql],
        capture_output=True, text=True, cwd=os.path.dirname(os.path.abspath(__file__))
    )
    if result.returncode != 0 and result.stderr:
        console.print(f"[bold red]Error:[/bold red] {result.stderr.strip()}")
    return result.stdout.strip()


def query_table(sql):
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
    return value.replace("\\", "\\\\").replace("'", "\\'")


def invalidate_user_permissions(user_id):
    result = subprocess.run(
        ["docker", "compose", "exec", "-T", "redis", "redis-cli",
         "-h", REDIS_HOST, "-p", REDIS_PORT,
         "DEL", f"user_permissions:{user_id}"],
        capture_output=True, text=True, cwd=os.path.dirname(os.path.abspath(__file__))
    )
    if result.returncode == 0 and result.stdout.strip() == "1":
        console.print(f"[dim]Cleared permission cache for user {user_id}[/dim]")


def print_table(title, rows, color="cyan"):
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


# ---------------------------------------------------------------------------
# Interactive pickers — arrow keys + enter
# ---------------------------------------------------------------------------

def render_pick_list(title, items, selected, multi=None):
    """Render a list of items with a cursor. `multi` is a set of selected indices for multi-select."""
    clear_screen()
    if title:
        console.print(Panel.fit(f"[bold white]{title}[/bold white]", border_style="cyan"))
    for i, item in enumerate(items):
        cursor = " > " if i == selected else "   "
        if multi is not None:
            checkbox = "[green][x][/green]" if i in multi else "[dim][ ][/dim]"
            console.print(f"{cursor}{checkbox} {item}")
        else:
            console.print(f"{cursor}{item}")
    if multi is not None:
        console.print(f"\n[dim]Arrow keys: navigate   Space: toggle   Enter: confirm   Esc: cancel[/dim]")
    else:
        console.print(f"\n[dim]Arrow keys: navigate   Enter: select   Esc: cancel[/dim]")


def pick_one(title, items, hint="Select"):
    """Arrow-key single selection. Returns index or None."""
    selected = 0
    while True:
        render_pick_list(title, items, selected)
        key = read_key()
        if key == "up":
            selected = (selected - 1) % len(items)
        elif key == "down":
            selected = (selected + 1) % len(items)
        elif key == "enter":
            return selected
        elif key == "escape":
            return None
        elif key == "ctrl_c":
            return None
        elif key.isdigit() and key != "0":
            idx = int(key) - 1
            if 0 <= idx < len(items):
                return idx
    return None


def pick_multi(title, items, initial=None):
    """Arrow-key multi-select with space to toggle. Returns set of selected indices or None."""
    selected = 0
    chosen = set(initial) if initial else set()
    while True:
        render_pick_list(title, items, selected, multi=chosen)
        key = read_key()
        if key == "up":
            selected = (selected - 1) % len(items)
        elif key == "down":
            selected = (selected + 1) % len(items)
        elif key == "enter":
            return chosen
        elif key == "escape":
            return None
        elif key == "ctrl_c":
            return None
        elif key == "space":
            if selected in chosen:
                chosen.discard(selected)
            else:
                chosen.add(selected)
        elif key.isdigit() and key != "0":
            idx = int(key) - 1
            if 0 <= idx < len(items):
                if idx in chosen:
                    chosen.discard(idx)
                else:
                    chosen.add(idx)
    return None


def confirm_choice(message, default=False):
    """Arrow-key Yes/No prompt. Returns bool."""
    options = ["No", "Yes"]
    selected = 1 if default else 0
    clear_screen()
    console.print(Panel.fit(f"[bold red]{message}[/bold red]", border_style="red"))
    while True:
        for i, label in enumerate(options):
            cursor = " > " if i == selected else "   "
            if i == selected:
                style = "bold green" if label == "Yes" else "bold red"
                console.print(f"{cursor}[{style}]{label}[/{style}]")
            else:
                console.print(f"{cursor}[dim]{label}[/dim]")
        console.print(f"\n[dim]Arrow keys: navigate   Enter: confirm[/dim]")
        key = read_key()
        if key == "up" or key == "left":
            selected = (selected - 1) % len(options)
            clear_screen()
            console.print(Panel.fit(f"[bold red]{message}[/bold red]", border_style="red"))
        elif key == "down" or key == "right":
            selected = (selected + 1) % len(options)
            clear_screen()
            console.print(Panel.fit(f"[bold red]{message}[/bold red]", border_style="red"))
        elif key == "enter":
            return selected == 1
        elif key == "escape":
            return False
        elif key == "ctrl_c":
            return False


# ---------------------------------------------------------------------------
# Plain commands
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
    rows = query_table(f"SELECT user_id FROM user_roles WHERE role_id = {role_id};")
    for row in rows:
        invalidate_user_permissions(row["user_id"])
    console.print(f"[bold green]Permissions set for role {role_id}.[/bold green]")


def cmd_assign_role(args):
    if len(args) < 2:
        console.print("[yellow]Usage: db.py assign-role <user_id> <role_id>[/yellow]")
        return
    user_id, role_id = args[0], args[1]
    run_sql(f"INSERT IGNORE INTO user_roles (user_id, role_id) VALUES ({user_id}, {role_id});")
    invalidate_user_permissions(user_id)
    console.print(f"[bold green]Role {role_id} assigned to user {user_id}.[/bold green]")


def cmd_remove_role(args):
    if len(args) < 2:
        console.print("[yellow]Usage: db.py remove-role <user_id> <role_id>[/yellow]")
        return
    user_id, role_id = args[0], args[1]
    run_sql(f"DELETE FROM user_roles WHERE user_id = {user_id} AND role_id = {role_id};")
    invalidate_user_permissions(user_id)
    console.print(f"[bold green]Role {role_id} removed from user {user_id}.[/bold green]")


def cmd_reset():
    if not confirm_choice("This will DROP ALL TABLES. Continue?"):
        console.print("[yellow]Aborted.[/yellow]")
        return
    console.print("[red]Dropping tables...[/red]")
    run_sql("SET FOREIGN_KEY_CHECKS=0; DROP TABLE IF EXISTS file_chunks, files, shared_links, drive_accounts, users, role_permissions, user_roles, permissions, roles; SET FOREIGN_KEY_CHECKS=1;")
    console.print("[bold green]All tables dropped.[/bold green] Restart the backend to recreate them.")
    console.print("[dim]Run: docker compose restart backend[/dim]")


# ---------------------------------------------------------------------------
# Interactive wrappers
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
    role_labels = [f"{r['name']} (id: {r['id']})" for r in roles]
    idx = pick_one(f"Step 1 — Choose a Role", role_labels)
    if idx is None:
        return
    role = roles[idx]

    perms = query_table("SELECT id, `key`, category FROM permissions ORDER BY category, `key`;")
    if not perms:
        console.print("[yellow]No permissions found.[/yellow]")
        return

    current_ids = query_table(f"SELECT permission_id FROM role_permissions WHERE role_id = {role['id']};")
    current_set = {i for row in current_ids for i in [int(row["permission_id"])]}

    perm_map = {p["id"]: p for p in perms}
    perm_ids_ordered = [p["id"] for p in perms]
    perm_labels = [f"[{p['category']}] {p['key']}" for p in perms]
    initial = [perm_ids_ordered.index(pid) for pid in current_set if pid in perm_ids_ordered]

    chosen_indices = pick_multi(f"Step 2 — Permissions for '{role['name']}'", perm_labels, initial)
    if chosen_indices is None:
        return
    if not chosen_indices:
        console.print("[yellow]No permissions selected. Aborted.[/yellow]")
        return
    perm_ids = [perm_ids_ordered[i] for i in chosen_indices]
    cmd_set_perm([role["id"]] + perm_ids)


def interactive_assign_role():
    users = query_table("SELECT id, email FROM users ORDER BY id;")
    if not users:
        console.print("[yellow]No users found.[/yellow]")
        return
    user_labels = [f"{u['email']} (id: {u['id']})" for u in users]
    uidx = pick_one("Step 1 — Choose a User", user_labels)
    if uidx is None:
        return
    user = users[uidx]

    roles = query_table("SELECT id, name FROM roles ORDER BY id;")
    if not roles:
        console.print("[yellow]No roles found.[/yellow]")
        return
    role_labels = [f"{r['name']} (id: {r['id']})" for r in roles]
    ridx = pick_one(f"Step 2 — Choose a Role for '{user['email']}'", role_labels)
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
    labels = [f"{a['email']} -> {a['role_name']}" for a in assignments]
    idx = pick_one("Select an assignment to remove", labels)
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


def main_menu():
    selected = 0
    while True:
        clear_screen()
        console.print(Panel.fit(
            f"[bold white]RouteStorage DB CLI[/bold white]\n[dim]{DB_NAME}@{DB_HOST}:{DB_PORT}[/dim]",
            border_style="cyan",
        ))
        for i, (label, _) in enumerate(MENU_ITEMS):
            cursor = " > " if i == selected else "   "
            if i == selected:
                console.print(f"{cursor}[bold white]{label}[/bold white]")
            else:
                console.print(f"{cursor}[dim]{label}[/dim]")
        console.print(f"\n[dim]Arrow keys: navigate   Enter/Space: select   q/Esc: quit[/dim]")

        key = read_key()
        if key in ("quit", "escape", "ctrl_c"):
            console.print("[dim]Bye.[/dim]")
            break
        elif key == "up":
            selected = (selected - 1) % len(MENU_ITEMS)
        elif key == "down":
            selected = (selected + 1) % len(MENU_ITEMS)
        elif key in ("enter", "space"):
            label, func = MENU_ITEMS[selected]
            clear_screen()
            console.print(Panel.fit(f"[bold white]{label}[/bold white]", border_style="green"))
            try:
                func()
            except Exception as e:
                console.print(f"[bold red]Error:[/bold red] {e}")
            console.input("\n[dim italic]Press Enter to continue...[/dim italic]")
        elif key.isdigit() and key != "0":
            idx = int(key) - 1
            if 0 <= idx < len(MENU_ITEMS):
                selected = idx
                label, func = MENU_ITEMS[selected]
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
