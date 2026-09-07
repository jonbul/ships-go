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

# NPCs (black holes, enemy ships, etc.)

NPC simulation (spawning, movement, lifecycle) no longer lives in this
service. It's owned by the sibling `ships-npc` project, which connects to
this server's `/ws` endpoint like a regular player, authenticates with the
`npcAuth` event (checked against `NPC_SECRET` + localhost), and pushes the
full current NPC batch on every tick via the `npcUpdate` event. ships-go just
stores the latest snapshot and relays it to players as `npcs` in the
`gameBroadcast` payload.

This server does add two small, generic pieces of plumbing for
**destructible** NPCs (currently used by `ships-npc`'s enemy Ship NPC, which
picks a random ship via `GET /game/getShips` and chases/shoots players):

- `npcHit` event (client -> server): a player reports one of its bullets hit
  an NPC (`npcId`, `bulletId`, `from`, `bulletCharge`, `x`, `y`). ships-go
  does no validation/damage math itself — it just forwards the message
  as-is to whichever connection is authenticated as the NPC controller
  (`isNpc == true`), which owns all NPC health/state.
- NPC death re-uses the existing `playerDied` event: `ships-npc` sends a
  `playerDied`-shaped message with `playerId` set to the NPC's own id when
  its health reaches zero, crediting the real killer via the normal kill-feed
  logic. No ships-go code change was needed for this because the existing
  handler never validated that the "victim" is a real connected player.
- Similarly, an NPC "shooting" a player reuses the existing `newBullet` /
  `removeBullet` / `playerHit` events untouched — `ships-npc` is simply
  another sender of `newBullet`, and clients self-detect being hit exactly
  like they do for other players' bullets.

# Prerequisites

- **Go** 1.26.2+
- **MongoDB** running instance (local or remote)
- A `.env` file with the required environment variables (connection string, etc.)

# Run

```
go run .
```