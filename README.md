# product-gallery

```
docker compose build
docker compose up -d
docker compose exec migrate sh -c "/scripts/migrate.sh"

docker compose down -t=1
```

```
docker compose build
docker compose up -d
docker compose exec importer sh -c "./importer"
```
