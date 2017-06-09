package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn/config"
	dataRedis "github.com/keratin/authn/data/redis"
	"github.com/keratin/authn/data/sqlite3"
	"github.com/keratin/authn/handlers"
)

func main() {
	cfg := config.ReadEnv()

	db, err := sqlite3.NewDB("dev")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	store := sqlite3.AccountStore{db}

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		panic(err)
	}
	redis := redis.NewClient(opts)

	tokenStore := dataRedis.RefreshTokenStore{
		Client: redis,
		TTL:    cfg.RefreshTokenTTL,
	}

	app := handlers.App{
		Db:                *db,
		Redis:             redis,
		Config:            cfg,
		AccountStore:      &store,
		RefreshTokenStore: &tokenStore,
	}

	r := mux.NewRouter()

	r.HandleFunc("/", app.Stub).Methods("GET")

	r.HandleFunc("/accounts", app.Stub).Methods("POST")
	r.HandleFunc("/accounts/import", app.Stub).Methods("POST")
	r.HandleFunc("/accounts/available", app.Stub).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", app.Stub).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", app.Stub).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", app.Stub).Methods("PUT", "PATCH")

	r.HandleFunc("/session", app.Stub).Methods("POST")
	r.HandleFunc("/session", app.Stub).Methods("DELETE")
	r.HandleFunc("/session/refresh", app.Stub).Methods("GET")

	r.HandleFunc("/password", app.Stub).Methods("POST")
	r.HandleFunc("/password/reset", app.Stub).Methods("GET")

	r.HandleFunc("/configuration", app.Stub).Methods("GET")
	r.HandleFunc("/jwks", app.Stub).Methods("GET")

	r.HandleFunc("/stats", app.Stub).Methods("GET")
	r.HandleFunc("/health", app.Health).Methods("GET")

	stack := gorilla.RecoveryHandler()(gorilla.CombinedLoggingHandler(os.Stdout, r))

	log.Fatal(http.ListenAndServe(":8000", stack))
}
