CHANGES
=======
Version 1.0.0 - 2026-XX-XX
------------------
- Initial release of the Ships game backend API in Go.
- Migrated from Express.js to Go for improved performance and scalability.
- Origin repository: https://github.com/jonbul/jaes
- Features:
  - User registration, login, and deletion.
  - Ship management (create, edit, return).
  - MongoDB integration for data storage.
  - WIP:
    - [X] WebSocket for real-time game functionality (WIP).
      - [X] Basic WebSocket implementation for game communication.
      - [X] Players moving and see each other in real-time.
      - [X] Background cards (stars)
      - [X] Bullets visible and hitting other players.
      - [X] Players dead and respawning.
      - [X] Check animation works in all users at same time.
    - [ ] Grafana dashboard for monitoring (WIP).