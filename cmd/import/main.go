package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // import postgres driver
)

type Product struct {
	ID          uint      `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Products []Product

func MockProducts() Products {
	return Products{
		{Name: "red", Description: "It's red", Slug: "red"},
		{Name: "blue", Description: "It's blue", Slug: "blue"},
		{Name: "green", Description: "It's green", Slug: "green"},
		{Name: "yellow", Description: "It's yellow", Slug: "yellow"},
		{Name: "brown", Description: "It's brown", Slug: "brown"},
		{Name: "orange", Description: "It's orange", Slug: "orange"},
	}
}

// CREATE TABLE IF NOT EXISTS products (
//     id SERIAL PRIMARY KEY,
//     name TEXT NOT NULL,
//     description TEXT,
//     slug VARCHAR(255) NOT NULL UNIQUE,
//     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
//     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
//     -- CONSTRAINT fk_author FOREIGN KEY(author_id) REFERENCES users(id)
// );

func main() {
	envCfg := envConfig()

	deadline := time.Now().Add(30 * time.Second)

	ctx := context.Background()
	ctx, ctxCancel := context.WithDeadline(ctx, deadline)
	_ = ctxCancel

	db, err := DbOpen(DbUrlBuilder(envCfg.dbUser, envCfg.dbPassword, envCfg.dbHost, envCfg.dbPort, envCfg.dbDbName))
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}

	for _, product := range MockProducts() {
		err = CreateProduct(ctx, db, &product)
		if err != nil {
			log.Fatalf("couldn't create product: %v", err)
		}
	}

	log.Println("√ done importing")
}

type config struct {
	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
	dbDbName   string
	// dbURI string
}

func envConfig() config {
	// dbURI, ok := os.LookupEnv("POSTGRESQL_URL")
	// if !ok {
	// 	panic("POSTGRESQL_URL not provided")
	// }

	dbUser, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		panic("POSTGRES_USER not provided")
	}
	dbPassword, ok := os.LookupEnv("PGPASSWORD")
	if !ok {
		panic("PGPASSWORD not provided")
	}
	dbHost, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		panic("POSTGRES_HOST not provided")
	}
	dbPort, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok {
		panic("POSTGRES_PORT not provided")
	}
	dbDbName, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		panic("POSTGRES_DB not provided")
	}

	return config{ /*dbURI*/ dbUser, dbPassword, dbHost, dbPort, dbDbName}
}

func DbUrlBuilder(bUser, dbPassword, dbHost, dbPort, dbDbName string) (url string) {
	// POSTGRESQL_URL="postgres://${POSTGRES_USER}:${PGPASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
	url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", bUser, dbPassword, dbHost, dbPort, dbDbName)
	return url
}

type DB struct {
	*sqlx.DB
}

func DbOpen(url string) (*DB, error) {
	db, err := sqlx.Open("postgres", url)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("√ successfully connected to database")
	return &DB{db}, nil
}

func CreateProduct(ctx context.Context, db *DB, product *Product) error {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := createProduct(ctx, tx, product); err != nil {
		return err
	}

	return tx.Commit()
}

func createProduct(ctx context.Context, tx *sqlx.Tx, product *Product) error {
	query := `
	INSERT INTO products (name, description, slug) 
	VALUES ($1, $2, $3) RETURNING id, created_at, updated_at
	`

	args := []interface{}{
		product.Name,
		product.Description,
		product.Slug,
	}

	err := tx.QueryRowxContext(ctx, query, args...).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("couldn't execute query (or scan?): %v", err)
	}

	return nil
}
