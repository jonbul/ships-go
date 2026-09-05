# About

Backend API for Ships game, a 2D multiplayer game with custom ships created by the user.
This is a project migrated from Express.js

- Frontend: https://github.com/jonbul/ships-vue
- Being migrated from: https://github.com/jonbul/jaes

# Use of AI

Only for audit

# TODO

Required for 1.0 and deploy

- [ ] Learn GO
- [X] Create a https Rest API with access to Mongo DB
  - [X] Register user, login and delete
  - [X] Manage ships (return, edit, create)
- [X] Websocket to make the game working as works now in https://jonbul.ddns.net

# For the Future
- Gravitational objects
- Animated objects
- Desktop controls udpate

# env

- `.env` lives in `../files/.env` (shared with `ships-npc`, and `ssl` certs
  in `../files/ssl` shared with `ships-vue`), symlinked into this project's
  root as `.env`/`ssl`. See `scripts/README.md` ("Shared config") for details
  and how `scripts/runDevEnvironment.sh` sets this up automatically.
- Required: `MONGODB_URI`
- `NPC_SECRET`: shared secret used to authenticate the `ships-npc` websocket
  connection (see `ships-npc`). NPC controller connections must also come
  from localhost.
- WIP

# NPCs (black holes, etc.)

NPC simulation (spawning, movement, lifecycle) no longer lives in this
service. It's owned by the sibling `ships-npc` project, which connects to
this server's `/ws` endpoint like a regular player, authenticates with the
`npcAuth` event (checked against `NPC_SECRET` + localhost), and pushes the
full current NPC batch on every tick via the `npcUpdate` event. ships-go just
stores the latest snapshot and relays it to players as `npcs` in the
`gameBroadcast` payload.

# Prerequisites

- **Go** 1.26.2+
- **MongoDB** running instance (local or remote)
- A `.env` file with the required environment variables (connection string, etc.)

# Run

```
go run .
```