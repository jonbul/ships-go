CHANGES
=======
Version 1.5.0 - 2026-09-XX
------------------
Support for the extracted ships-npc service, a tabbed dev log viewer, and a
round of concurrency fixes in the websocket hub.

- Only one NPC controller is accepted at a time. Every `npcUpdate` replaces
  the whole NPC snapshot, so two connected controllers did not add up - they
  overwrote each other every tick, and players saw an enemy ship flicker in
  and out of existence: unnamed, undrawn and impossible to hit, while its
  bullets kept arriving. The first to authenticate keeps the role until it
  disconnects; a duplicate is refused with an `npcRejected` event explaining
  why. Refusing the newcomer is deliberate - taking over instead would make
  two live controllers evict each other in a loop.
- When the NPC controller disconnects its NPCs are dropped. Keeping the last
  snapshot stranded ghost ships that never moved and could never be killed.
- `NpcData` carries `kills`/`deaths` so ships-vue can list NPC ships in the
  scoreboard next to the players.
- New admin setting: enemy ships attack each other. Stored and pushed to
  ships-npc like every other NPC setting, with no new endpoint.
- Player broadcasts now carry the player's raw ship width/height. A shipId
  alone cannot identify a ship to anyone else, because GET /game/getShips
  lists only the public ones; ships-npc needs the real numbers to aim at the
  middle of a player's ship rather than at a guess. Passthrough only - see
  ships-npc/CHANGES.md 1.0.0.
- New `scripts/viewLogs.sh`: a tabbed log viewer for the dev environment.
  `runDevEnvironment.sh` used to interleave the ships-go, ships-npc and
  ships-vue logs into a single unreadable stream. Now one tab per service is
  shown, with a pinned tab bar on the top line, plus an "All" tab that merges
  the three with a coloured prefix per service. Keys: `1`-`9`, `Tab`/arrows,
  `a` all, `c` clear, `r` reload, `d` detach (leaves the services running),
  `q` quit. It is plain bash and ANSI escapes, so it needs no tmux/screen,
  and it can be run standalone to reattach to a running dev environment. It
  is the default in every terminal, Konsole included, so the dev environment
  behaves the same in a plain console, over SSH and in VS Code;
  `--konsole-tabs` still opens one real Konsole tab per service and
  `--no-ui` restores the plain interleaved output.
- The tab bar stays put: it is repainted twice a second (a log line carrying
  a stray escape sequence could leave it stale or blank), the scrolling
  region is re-asserted on every repaint, and the viewer runs on the
  alternate screen so the terminal's own scrollback cannot drag the bar out
  of view. Scrollback is handled by the viewer rather than the terminal
  (Up/Down, PgUp/PgDn, Home, End/f/G), and scrolling away pauses the tail so
  the text stops moving while you read it. The bar says whether you are
  watching live output or a frozen window.
- First automated tests for the websocket package, covering the broadcast
  drain, the idle heartbeat and the background build race.

Bugfixes
- Memory grew without limit while nobody was playing. ships-npc stays
  connected and keeps simulating, so bullets and kills kept arriving, but the
  broadcast loop returned early when the player list was empty - before the
  point where those buffers are drained. They were held until somebody joined
  and then delivered as one enormous frame. The buffers are now always
  drained, and an empty broadcast still goes out every two seconds so
  ships-npc learns the last player left instead of chasing and shooting a
  ghost forever.
- One slow or unresponsive client could freeze the game for everyone. Every
  broadcast was written to every socket while holding the global game lock,
  with no write deadline, so a single stalled connection blocked the next
  tick. Payloads are now built under the lock and written after it is
  released, every write has a five second deadline, and a failed write closes
  the connection so the usual disconnect cleanup runs instead of leaving a
  dead socket in the game forever. `npcHit` forwarding follows the same rule.
- The admin panel and /game/getPlayers could kill the server. Both serialised
  the live player map straight out of the websocket package while player
  goroutines were writing to it, which Go aborts the whole process for
  ("concurrent map iteration and map write") - it is not a recoverable panic.
  They now read a snapshot taken under the game lock.
- The star background was built lazily behind a length check from every
  client's connection goroutine, so two players joining at the same moment
  could crash the server the same unrecoverable way, or see a half-built map.
  It is now built exactly once.
- The NPC controller flag is now read and written atomically. It was written
  under one lock and read under another (and in one place under none), which
  is not synchronisation at all.
- Killing a player now queues the killer's updated credit total for the next
  broadcast, instead of relying on that player happening to send an update.
- A player's `hide` flag never reached anyone. It was tagged `hidden` on the
  way in and out, and no client has ever sent or read that name, so it
  decoded as `false` every time: a player waiting out their respawn stayed
  drawn on everybody else's screen for the full ten seconds.
- Security: the NPC secret is compared in constant time.
- `runDevEnvironment.sh` left services running after being stopped. `go run`
  compiles to `/tmp/go-build.../exe/<name>` and execs it as a child, so
  killing the `go run` pid orphaned the actual server - which is how two
  ships-npc controllers ended up connected at once. Services are now started
  with `setsid` and the whole process group is killed on exit.
- The dev script's cleanup handler aborted half-way and exited with 143.
  `wait` returns 143 for a child killed by SIGTERM and `set -e` treats that
  as a failure; both `wait` calls are now guarded. With `--konsole-tabs` no
  background pids are recorded, so `wait "${PIDS[@]}"` expanded to a bare
  `wait`, which waits for *every* child - i.e. the Konsole tabs - and never
  returned, hanging the script on exit.
- Terminal size was read with `tput`, which prefers an inherited
  LINES/COLUMNS pair over asking the terminal, so the tab bar and scrolling
  region kept using the old size after a resize. Read from `stty size`
  instead. A key pressed while the viewer was repainting was also echoed into
  the log view as literal noise (`^[[A` for an arrow key); echo is now off
  for the session.
- Removed a leftover copy of the dev environment script and a flag that
  nothing ever read.

Version 1.4.1 - 2026-09-05
------------------
- NPC enemy ship speed is now carried as a game-unit speed
  (`enemyShipSpeed`, same scale as ships-vue's `SPEED.MAX` of 50) rather
  than a 0-1 fraction, and the built-in defaults were updated to match
  ships-npc: 1 ship, 10 life, speed 20, 500 ms fire rate, 2 black holes,
  30 s spawn period.

Version 1.4.0 - 2026-09-05
------------------
- New `npcConfig` websocket event and NPC settings storage, so an
  administrator can tune the NPCs from ships-vue's admin panel while the
  game is running. The settings are exposed through the existing
  `GET /game/admin/data` and `POST /game/admin` endpoints and pushed to
  the authenticated `ships-npc` controller connection immediately on save,
  and again right after it authenticates - so whichever process restarted
  last converges on the same values, with no polling and no second
  endpoint. The browser never talks to `ships-npc` directly.
- Values are clamped server-side (`NpcSettingsData.Sanitized`) and echoed
  back in the save response, so the admin panel shows what is genuinely in
  force rather than a number that was silently rejected downstream.
  `ships-npc` clamps again on receipt, since it can't trust the wire.

Version 1.3.0 - 2026-09-05
------------------
- New `npcHit` websocket event: forwards a player's report of a bullet
  hitting an NPC straight to the `ships-npc` controller connection (no
  validation/damage logic here, NPC keeps owning its own health/state).
- Supports destructible NPCs (used by `ships-npc`'s new enemy Ship NPC)
  by reusing the existing `playerDied`/`newBullet`/`removeBullet`/
  `playerHit` events untouched - no other backend changes were required.

Version 1.2.0 - 2026-09-XX
------------------
- NPC simulation (black holes) moved out to the new `ships-npc` service.
  ships-go no longer spawns/moves NPCs itself: it relays whatever
  `ships-npc` sends via the new `npcAuth`/`npcUpdate` websocket events.
- `gameBroadcast` payload field renamed `blackHoles` -> `npcs` (generic,
  ready for multiple NPC kinds/instances).
- New `NPC_SECRET` env var required to authenticate NPC controller
  connections (also restricted to localhost).

Version 1.1.0 - 2026-09-04
------------------
- New black hole managed from backend

Version 1.0.3 - 2026-08-09
------------------
- Missing borderWidth in shape Struct

Version 1.0.2 - 2026-08-06
------------------
- Concurrent users bug

Version 1.0.1 - 2026-08-04
------------------
- Disconnected users aren't removed propperly

Version 1.0.0 - 2026-08-02
------------------
- Initial release of the Ships game backend API in Go.
- Migrated from Express.js to Go for improved performance and scalability.
- Origin repository: https://github.com/jonbul/jaes
- Features:
  - User registration, login, and deletion.
  - Ship management (create, edit, return).
  - MongoDB integration for data storage.
  - WIP:
    - [X] WebSocket for real-time game functionality.
      - [X] Basic WebSocket implementation for game communication.
      - [X] Players moving and see each other in real-time.
      - [X] Background cards (stars)
      - [X] Bullets visible and hitting other players.
      - [X] Players dead and respawning.
      - [X] Check animation works in all users at same time.
    - [X] Monitoring.
  - BUGS:
    - [X] Disconnected users aren't removed propperly
    