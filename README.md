# About

Backend API for Ships game, a 2D multiplayer game with custom ships created by the user.
This is a project migrated from Express.js

- Frontend: https://github.com/jonbul/ships-vue
- Being migrated from: https://github.com/jonbul/jaes

# Use of AI

Only as audit

# TODO

Required for 1.0 and deploy

- [ ] Learn GO
- [ ] Create a https Rest API with access to Mongo DB
  - [ ] Register user, login and delete
  - [ ] Manage ships (return, edit, create)
- [ ] Websocket to make the game working as works now in https://jonbul.ddns.net

# env

- Create a `.env` file in the project root
- Required: `MONGODB_URI`
- WIP

# Prerequisites

- **Go** 1.26.2+
- **MongoDB** running instance (local or remote)
- A `.env` file with the required environment variables (connection string, etc.)

# Run

```
go run .
```