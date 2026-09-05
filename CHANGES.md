CHANGES
=======
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
    