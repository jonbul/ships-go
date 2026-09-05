CHANGES
=======
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
    