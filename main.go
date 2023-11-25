package main

import (
	//...

	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // import postgres driver
)

type config struct {
	port string

	dbHost     string
	dbPort     string
	dbUser     string
	dbPassword string
	dbDbName   string
}

func main() {
	envCfg := envConfig()

	deadline := time.Now().Add(30 * time.Second)

	ctx := context.Background()
	ctx, ctxCancel := context.WithDeadline(ctx, deadline)
	_ = ctxCancel

	db, err := DbOpen(DbUrlBuilder(envCfg.dbHost, envCfg.dbPort, envCfg.dbUser, envCfg.dbPassword, envCfg.dbDbName))
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	fs := http.FileServer(http.Dir("/static"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))
	r.Get("/static/*", http.StripPrefix("/static/", fs).ServeHTTP)
	// r.Get("/static/", fs.ServeHTTP)

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("hi"))
	// })
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("./templates/index.html")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Write(data)
	})

	// create productsCtx

	// RESTy routes for "products" resource
	r.Route("/products", func(r chi.Router) {
		// r.Use(ProductsCtx)            // Load the *Product on the request context
		r.Get("/", listProducts(ctx, db)) // GET /products
		// r.With(paginate).Get("/", listProducts) // GET /products
	})

	http.ListenAndServe(fmt.Sprintf(":%s", envCfg.port), r)
}

func envConfig() config {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		panic("PORT not provided")
	}

	dbHost, ok := os.LookupEnv("POSTGRES_HOST")
	if !ok {
		panic("POSTGRES_HOST not provided")
	}
	dbPort, ok := os.LookupEnv("POSTGRES_PORT")
	if !ok {
		panic("POSTGRES_PORT not provided")
	}
	dbUser, ok := os.LookupEnv("POSTGRES_USER")
	if !ok {
		panic("POSTGRES_USER not provided")
	}
	dbPassword, ok := os.LookupEnv("PGPASSWORD")
	if !ok {
		panic("PGPASSWORD not provided")
	}
	dbDbName, ok := os.LookupEnv("POSTGRES_DB")
	if !ok {
		panic("POSTGRES_DB not provided")
	}

	return config{ /*dbURI*/ port, dbHost, dbPort, dbUser, dbPassword, dbDbName}
}

func DbUrlBuilder(dbHost, dbPort, dbUser, dbPassword, dbDbName string) (url string) {
	// POSTGRESQL_URL="postgres://${POSTGRES_USER}:${PGPASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"
	url = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbDbName)
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

	log.Println("successfully connected to database")
	return &DB{db}, nil
}

type Index struct{}

// paginate is a stub, but very possible to implement middleware logic
// to handle the request params for handling a paginated request.
func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}

type Product struct {
	ID          uint      `json:"-"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type Products []Product

func listProducts(ctx context.Context, db *DB) http.HandlerFunc {
	tx, err := db.BeginTxx(ctx, nil)

	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	defer tx.Rollback()

	products, err := findProducts(ctx, tx, ProductFilter{})

	// return func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte(fmt.Sprintf("products: %s", products)))
	// }
	return func(w http.ResponseWriter, r *http.Request) {

		tmplt := template.New("product-gallery.html.tmpl")                     //create a new template with some name
		tmplt, err = tmplt.ParseFiles("./templates/product-gallery.html.tmpl") //parse some content and generate a template, which is an internal representation
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tProducts := Products{}
		for _, product := range products {
			tProducts = append(tProducts, Product{product.ID, product.Name, product.Description, product.Slug, product.CreatedAt, product.UpdatedAt})
		}

		type TmplData struct {
			Products []Product
		}
		tmplData := TmplData{Products: tProducts}

		err = tmplt.Execute(w, tmplData) //merge template ‘t’ with content of ‘p’
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type ProductFilter struct {
	ID          *uint
	Name        *string
	Description *string
	Slug        *string

	Limit  int
	Offset int
}

func findProducts(ctx context.Context, tx *sqlx.Tx, filter ProductFilter) ([]*Product, error) {
	where, args := []string{}, []interface{}{}
	argPosition := 0 // used to set correct postgres argument enums i.e $1, $2

	if v := filter.ID; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("id = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Slug; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("slug = $%d", argPosition)), append(args, *v)
	}

	if v := filter.Name; v != nil {
		argPosition++
		where, args = append(where, fmt.Sprintf("name = $%d", argPosition)), append(args, *v)
	}

	query := "SELECT * from products" + formatWhereClause(where) + " ORDER BY created_at DESC"
	log.Printf("query: %s", query)
	products, err := queryProducts(ctx, tx, query, args...)
	log.Printf("products (from queryProducts): %v", products)

	if err != nil {
		return products, err
	}

	return products, nil
}

func formatWhereClause(where []string) string {
	if len(where) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(where, " AND ")
}

func queryProducts(ctx context.Context, tx *sqlx.Tx, query string, args ...interface{}) ([]*Product, error) {
	products := make([]*Product, 0)
	err := findMany(ctx, tx, &products, query, args...)

	if err != nil {
		return products, err
	}

	return products, nil
}

func findMany(ctx context.Context, tx *sqlx.Tx, ss interface{}, query string, args ...interface{}) error {
	rows, err := tx.QueryxContext(ctx, query, args...)

	if err != nil {
		return err
	}

	defer rows.Close()

	sPtrVal, err := asSlicePtrValue(ss) // get the reflect.Value of the ptr to slice

	if err != nil {
		return err
	}

	sVal := sPtrVal.Elem()                           // get the relfect.Value of the slice pointed to by ss
	newSlice := reflect.MakeSlice(sVal.Type(), 0, 0) // new slice
	elemType := sliceElemType(sVal)                  // get the slice element's type

	for rows.Next() {
		newVal := reflect.New(elemType) // create a new value of this type
		if err := rows.StructScan(newVal.Interface()); err != nil {
			return nil
		}
		newSlice = reflect.Append(newSlice, newVal)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	sPtrVal.Elem().Set(newSlice) // change the value pointed to be the ptr to slice to our new slice

	return nil
}

func asSlicePtrValue(v interface{}) (reflect.Value, error) {
	if !isSlicePtr(v) {
		return reflect.Value{}, errors.New("expecting a pointer to slice")
	}
	return reflect.ValueOf(v), nil
}

// sliceElemType takes a reflect.Value which is a ptr to slice or a slice,
// and returns the reflect.Type of the elements the slice holds.
// If the slice holds a pointer type, it returns the type pointed to.
func sliceElemType(v reflect.Value) reflect.Type {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	vv := v.Type().Elem() // get the reflect.Type of the elements of the slice

	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem() // if it is a pointer, get the type it points to
	}

	return vv
}

func isSlicePtr(v interface{}) bool {
	typ := reflect.TypeOf(v)

	return typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Slice
}
