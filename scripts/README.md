# MongoDB Container

This is a simple MongoDB container for development purposes.
It uses the official MongoDB 7.0 image from Docker Hub.

## Requirements

- Docker installed (the script will install it automatically if not found on Fedora/RHEL systems)

## First time setup

The first time you run the script it will check if Docker is installed.
If it's not installed, it will install it and ask you to **log out and log back in** before running the script again.

## Usage

Run the following command from the `db/` directory:

```bash
bash runMongoContainer.sh
```

- If a container named `mongodb` already exists (running or stopped), the script will ask whether to recreate it.
- On startup, the script will automatically:
  - Create an `admin` root user
  - Create a `testAdmin` user with full privileges
  - Create a `test` user with read/write access to the `jaes` database
  - Create collections and load test data from `./testData/` (one file per collection, named `{database}.{collection}.json`)

- The script will stay running until you type `exit`, at which point the container will be **stopped and removed**.

## Credentials

| User | Password | Role |
|------|----------|------|
| `admin` | `admin` | Root (all databases) |
| `testAdmin` | `testAdmin` | Read/Write/Admin (all databases) |
| `test` | `test` | Read/Write (`jaes` database) |

## Connection

### Client

```
mongodb://testAdmin:testAdmin@127.0.0.1:27017/?authSource=admin&directConnection=true
```
### API (.env)
```
mongodb://test:test@127.0.0.1:27017/?authSource=jaes&directConnection=true
```


## Test data

Test data is loaded from JSON files in `./testData/` following the naming convention:

```
{database}.{collection}.json
```

# Full dev environment

`runDevEnvironment.sh` launches everything needed for local development in
one command: the MongoDB container (`runMongoContainer.sh`), the `ships-go`
backend, the `ships-npc` NPC controller and the `ships-vue` frontend
(`runWeb.sh`).

```bash
bash runDevEnvironment.sh
```

- Run it from a real terminal: MongoDB's script is interactive (it may ask
  to recreate an existing container, and waits for you to type `exit`).
- Shared config first: `.env` and SSL certs are common to more than one
  project, so they live once in `files/` (a sibling of `ships-go`,
  `ships-npc` and `ships-vue`, **outside** every project folder), and each
  project links to them — see [Shared config (`files/`)](#shared-config-files)
  below. Before starting anything, this script:
  - Generates `files/ssl/cert.pem` / `key.pem` (self-signed, `localhost`) if
    they don't exist yet.
  - Recreates the `ships-go/.env`, `ships-npc/.env`, `ships-go/ssl` and
    `ships-vue/ssl` symlinks if missing (e.g. right after a fresh clone).
  - It does **not** create `files/.env` itself — that file holds secrets
    (`MONGODB_URI`, `NPC_SECRET`, ...) you must provide manually the first
    time; the script prints a warning if it's missing.
- It then runs `runMongoContainer.sh` and waits for MongoDB to accept
  connections before starting `ships-go`, `ships-npc` and `ships-vue`.
- Logs: `ships-go`, `ships-npc` and `ships-vue` run in the background and
  the script opens the **tabbed log viewer** described below, so the three
  logs stay separated instead of being interleaved in one stream. This is
  now the default in every terminal — a plain console, over SSH or VS Code's
  integrated terminal — so the dev environment always behaves the same way.
  Everything is still written to `scripts/logs/<name>.log`, so `tail -f`
  keeps working too. Inside Konsole you can still pass `--konsole-tabs` to
  get the old one-real-tab-per-service behaviour.
- Press Ctrl+C to stop whatever this script started in the background
  (non-Konsole mode). The MongoDB container is left running either way.

## Tabbed log viewer (`viewLogs.sh`)

`viewLogs.sh` shows one log per tab in a single terminal, with a tab bar on
the top line and the selected log scrolling underneath. It's plain bash +
ANSI escapes: no `tmux`, `screen` or any other dependency.

`runDevEnvironment.sh` opens it automatically. You can also run it on its own
at any time — including while the dev environment is already running, to
reattach after detaching — since it only reads the log files:

```bash
bash viewLogs.sh                     # reads scripts/logs/
bash viewLogs.sh --log-dir /some/dir # any directory of .log files
```

The tab bar always stays on the top line: row 1 is kept outside the
terminal's scrolling region, it is repainted twice a second so a stray
escape sequence in a log line can't leave it stale, and the viewer runs on
the alternate screen so the terminal's own scrollback can't drag it out of
view. To look back through a log, use the keys below instead — the bar stays
put and tells you whether you're watching live output (`LIVE`) or a frozen
window (`SCROLLED -n`).

| Key | Action |
|-----|--------|
| `1`-`9` | Jump to that tab |
| `Tab` / `→` | Next tab |
| `Shift+Tab` / `←` | Previous tab |
| `a` | **All** tab: the three logs merged, each line prefixed with a coloured service name |
| `↑` / `↓` | Scroll back/forward one line (pauses following) |
| `PgUp` / `PgDn` | Scroll a page |
| `Home` | Oldest line |
| `End` / `f` / `G` | Back to the live tail |
| `c` | Clear the screen (the log file itself is untouched) |
| `r` | Reload the current log from the start |
| `d` | **Detach**: leave the viewer but keep the services running |
| `q` | Quit; when launched by `runDevEnvironment.sh` this also stops the services |

Pass `--no-ui` to `runDevEnvironment.sh` if you'd rather have the old plain
interleaved output.

## Shared config (`files/`)

`ships-go`, `ships-npc` and `ships-vue` are separate git repos, but some
config is shared between them and shouldn't be duplicated or committed:

- `.env` — `ships-go` and `ships-npc` need the *same* `NPC_SECRET`; keeping
  one file avoids them drifting out of sync.
- `ssl/cert.pem` + `ssl/key.pem` — `ships-go` (Gin `RunTLS`) and `ships-vue`
  (Vite HTTPS dev server) both need a certificate for local HTTPS; keeping
  one avoids duplicated, identical cert files in two repos.

Both live in `files/` (`/home/jonander/git/ships/files/`, next to `ships-go`,
`ships-npc` and `ships-vue`), and each project has a **symlink** pointing at
them: `ships-go/.env`, `ships-npc/.env`, `ships-go/ssl`, `ships-vue/ssl` all
resolve to `../files/.env` / `../files/ssl`. This is transparent to each
project's code (Go's `godotenv.Load()` and Vite's `fs.readFileSync` just see
a normal `.env` file / `ssl` folder in their own directory) — no code
changes needed.

Because `files/` lives outside every repo, these symlinks are gitignored and
won't exist right after a fresh clone. `runDevEnvironment.sh` recreates them
automatically (see above); if you're not using that script, create them
yourself, e.g.:

```bash
ln -s ../files/.env ships-go/.env
ln -s ../files/.env ships-npc/.env
ln -s ../files/ssl  ships-go/ssl
ln -s ../files/ssl  ships-vue/ssl
```

`files/.env` needs at least: `MONGODB_URI`, `NPC_SECRET`,
`SSL_CERT_PATH=./ssl/cert.pem`, `SSL_KEY_PATH=./ssl/key.pem` (see
`ships-go/README.md` for the full list).

## VS Code alternative: one tab per project

If you work from VS Code, `.vscode/tasks.json` (in `ships-go`) defines a task
per project. Running a task opens it in its own terminal tab in the Panel,
giving you the same "one tab per log" experience natively:

1. Open the `ships-go` folder in VS Code.
2. `Terminal > Run Task... > MongoDB` (interactive, run once).
3. Once MongoDB is up, `Terminal > Run Task... > Dev: ships-go + ships-npc + ships-vue`
   to start the other 3 in parallel, each in its own tab.


