# product-gallery

## about
this app:
* is meant as a simple app example for an deployment with some extras like db,migrations... etc
* uses solemly containers (meant for cloud-native deployments)
* should be easy updatable for showcasing smth like canary or blue/green deployment
* uses a db to showcase / include db migrations + beeing dependend on another service
* has an importer service for an inital product data import

## setup:
``` shell
# build + start all
docker compose build
docker compose up -d

# run migrations (via the migrations service, see cicd/container-images/migrate/migrate.sh)
docker compose exec migrate sh -c "/scripts/migrate.sh"

# run importer (via importer service, see: cmd/import/main.go & cicd/container-images/app/Dockerfile:12 )
docker compose exec importer sh -c "./importer"

# visit or curl curl -v http://127.0.0.1:3333/products
curl -v http://127.0.0.1:3333/products
```

## for viewing db/postgres we got an adminer
* visit http://127.0.0.1:8080
* see [docker-compose.yml](docker-compose.yml) for `username` / `password` / etc

## services explained (docker-compose services)
* "postgres" => the app database
* "adminer" => an extra tool to view the database content voa a webbrowser (after startup visit http://127.0.0.1:8080)
* "migrate" => service which includes a migration tool to run database migrations
  * see: https://github.com/golang-migrate/migrate
  * runs with `sleep` command to provide an easy way to run migrations via cli via docker exec
* "importer" => service which includes a binary which runs a simple init data import against the database
  * runs with `sleep` command to provide an easy way to run the import via cli via docker exec
* "app" => contains & runs the application
