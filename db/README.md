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
