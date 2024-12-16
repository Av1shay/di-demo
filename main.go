package main

import (
	"context"
	"github.com/Av1shay/di-demo/cache/memory"
	"github.com/Av1shay/di-demo/cache/redis"
	"github.com/Av1shay/di-demo/repositories/mongo"
	"github.com/Av1shay/di-demo/repositories/mysql"
	"github.com/Av1shay/di-demo/server"
	"github.com/Av1shay/di-demo/uam"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	defaultPort = "8050"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	serv := resolveDependencies(ctx)

	serv.MountHandlers()

	log.Println("Server listening on port " + port)

	if err := http.ListenAndServe(":"+port, serv.Router()); err != nil {
		log.Fatal("Server failed to listen:", err)
	}
}

func resolveDependencies(ctx context.Context) *server.Server {
	var (
		itemRepo uam.ItemRepository
		cache    uam.Cache
		err      error
	)

	switch ds := os.Getenv("DATA_SOURCE"); ds {
	case "mysql":
		itemRepo, err = mysql.NewRepository(os.Getenv("MYSQL_CONNECTION"))
		if err != nil {
			log.Fatal("Error creating items repository", err)
		}
	case "mongo":
		itemRepo, err = mongo.NewRepository(os.Getenv("MONGO_URI"), os.Getenv("MONGO_DB"))
		if err != nil {
			log.Fatal("Error creating new mysql user repository", err)
		}
	default:
		log.Fatalf("Invalid or empty DATA_SOURCE config provided: %q", ds)
	}

	if os.Getenv("CACHE_PROVIDER") == "redis" {
		rdb := redissdk.NewClient(&redissdk.Options{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASSWORD"),
		})
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Println("Failed to ping Redis", err)
		}
		cache = redis.NewCache(rdb)
	} else {
		// default in-memory cache
		cache = memory.NewCache()
	}

	cacheEnabled := false
	if ce, err := strconv.ParseBool(os.Getenv("CACHE_ENABLED")); err == nil {
		cacheEnabled = ce
	}

	uamAPI, err := uam.NewAPI(uam.Config{CacheEnabled: cacheEnabled}, itemRepo, cache)
	if err != nil {
		log.Fatal("Error creating UAM API", err)
	}

	return server.New(uamAPI)
}
