# Quarm DB Backup

Silly little container to back up a Quarm database on an interval.

## Usage

```
git clone https://github.com/DDTully/quarm-backup.git
cd quarm-backup
cp .env.example .env
```

Edit the `.env` file to set your backup preferences.

Docker compose is required to run the backup service. If you don't have it installed, follow the instructions at <https://docs.docker.com/compose/install/>.

```
docker compose up -d
```

SQL dumps will be created in the `database_backups` directory.
