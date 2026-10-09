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
- New admin settings for the two NPC fleets: which controllers run (neither,
  the rule-based one, the AI one, or both), a separate size for each fleet -
  now up to 100 ships each - and a six-cell matrix saying which fleet may
  attack players, rule ships and AI ships. Stored and pushed to ships-npc
  like every other NPC setting, with no new endpoint. They are validated on
  the way in: an unrecognised controller falls back to the default rather
  than silently meaning "no NPCs at all".
- Black hole duration is an admin setting (5s to a day, default the 180s it
  was fixed at), alongside the existing cap and spawn period.
- New `gameSettings` event, sent to each player on connection and broadcast
  again whenever an admin saves. It carries the rules a browser has to
  enforce for itself: whether ships take damage on contact, whether a ship
  grows with its score, and the standard size every ship is drawn at (100 by
  default, the value ships-vue used to hardcode). All three are resolved
  client-side and so cannot be switched off in one place. The ship size goes
  to ships-npc as well, since it decides how big the ships it is flying
  actually are; it is clamped to 20-1000, and an unset value becomes the
  default rather than the minimum, which would shrink every ship.
  Sending it live rather than at page load means the change reaches players
  already in the game.
- Ship life is one setting for everybody, renamed from `enemyShipLife` to
  `shipLife` and carried in `gameSettings` as well as `npcConfig`. It was an
  NPC-only knob, so raising it armoured every NPC while leaving the player on
  the 10 ships-vue had hardcoded - the opposite of the "same rules for
  everyone" the other game rules follow. Nothing already flying is healed or
  hurt by a change; it applies from each ship's next spawn, which is how the
  NPC side has always read it.
- Player broadcasts now carry the player's raw ship width/height. A shipId
  alone cannot identify a ship to anyone else, because GET /game/getShips
  lists only the public ones; ships-npc needs the real numbers to aim at the
  middle of a player's ship rather than at a guess. Passthrough only - see
  ships-npc/CHANGES.md 1.0.0.
- **ships-npc's CPU and memory now appear on `/metrics`**, alongside the Go
  and process collectors that describe this server. The NPC simulation is the
  most computationally expensive thing in the game, and since it moved out
  into its own service it was the one part with no monitoring at all: a fleet
  of 200 ships could be pushing a core flat and nothing said so.

  ships-npc has no HTTP server, and as a websocket client on loopback it is
  not reachable to be scraped, so it *pushes* a sample up the connection it
  already has (new `npcMetrics` event, accepted only from the authenticated
  NPC controller) and this server re-exports it. No new port, listener,
  credential or Prometheus target - the existing scrape picks it up, so the
  Grafana side is a new panel rather than a new job. Series:
  `ships_npc_cpu_seconds_total` (a counter, for `rate()`),
  `ships_npc_cpu_percent`, `ships_npc_memory_resident_bytes`,
  `ships_npc_memory_heap_bytes`, `ships_npc_memory_heap_sys_bytes`,
  `ships_npc_goroutines`, `ships_npc_simulated_npcs`,
  `ships_npc_tick_seconds` and `ships_npc_up`.

  Collected on demand rather than kept in gauges the handler writes to.
  Gauges hold their last value forever, so a dead ships-npc would leave the
  dashboard showing a healthy service frozen at a plausible CPU figure;
  instead `ships_npc_up` drops to 0 - immediately when the controller
  disconnects, or after 30s of silence if it stops reporting without closing
  the socket - and the other series stop being reported at all.
- New Grafana dashboard for all of the above:
  `scripts/monitoringContainers/shipsDashboard.json`, importable as it is.
  Twenty-one panels in five rows: an overview strip, resource usage,
  simulation load, capacity and the Go runtime.

  The capacity row is the point of it. It divides CPU and tick time by the
  number of NPCs alive, so the cost of *one* ship can be read off the graph
  and multiplied by a fleet size to predict the load before that fleet is
  actually flown. A flat line means the simulation scales linearly; a rising
  one means something in it is quadratic, which ship-vs-ship collision and
  targeting naturally are. Tick duration is drawn against a `tick_budget_ms`
  variable, since a tick time only means something next to the interval it
  has to fit inside.

  Every panel carries a description explaining what it is for and how to read
  it, so the reasoning shows up in Grafana's own tooltips rather than only in
  this file. Exported in the shareable `__inputs` format, so importing
  prompts for the Prometheus datasource instead of carrying a UID from
  whichever Grafana it was built against.
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
    