package main

import (
	"context"
	"github.com/Av1shay/di-demo/authentication"
	"github.com/Av1shay/di-demo/cache/memory"
	"github.com/Av1shay/di-demo/cache/redis"
	"github.com/Av1shay/di-demo/config"
	"github.com/Av1shay/di-demo/repositories/mongo"
	"github.com/Av1shay/di-demo/repositories/mysql"
	"github.com/Av1shay/di-demo/server"
	"github.com/Av1shay/di-demo/uam"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

const (
	defaultPort = "8050"
)

func init() {
	logLVL := slog.LevelInfo
	if debug, _ := strconv.ParseBool(os.Getenv("DEBUG")); debug {
		logLVL = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLVL})
	slog.SetDefault(slog.New(h))
}

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

	confManager := config.NewManager()
	if err := confManager.Validate(); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	serv := resolveDependencies(ctx, confManager)

	serv.MountHandlers()

	log.Println("Server listening on port " + port)

	if err := http.ListenAndServe(":"+port, serv.Router()); err != nil {
		log.Fatal("Server failed to listen:", err)
	}
}

func resolveDependencies(ctx context.Context, confManager *config.Manager) *server.Server {
	var (
		repo  uam.Repository
		cache uam.Cache
		err   error
	)

	switch confManager.DataSource() {
	case config.DataSourceMySQL:
		repo, err = mysql.NewRepository(confManager.MySQLConn())
		if err != nil {
			log.Fatal("Error creating mysql repository: ", err)
		}
	case config.DataSourceMongo:
		repo, err = mongo.NewRepository(confManager.MongoURI(), confManager.MongoDB())
		if err != nil {
			log.Fatal("Error creating new mongo repository: ", err)
		}
	}

	if confManager.CacheProvider() == config.CacheProviderRedis {
		rdb := redissdk.NewClient(&redissdk.Options{
			Addr:     confManager.RedisAddr(),
			Password: confManager.RedisPassword(),
		})
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Println("Failed to ping Redis: ", err)
		}
		cache = redis.NewCache(rdb)
	} else {
		// default in-memory cache
		cache = memory.NewCache()
	}

	uamAPI, err := uam.NewAPI(confManager.UAMAPIConfig(), repo, cache)
	if err != nil {
		log.Fatal("Error creating UAM API: ", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	auth := authentication.NewClient()

	serv, err := server.New(ctx, validate, auth, uamAPI)
	if err != nil {
		log.Fatal("Error creating server: ", err)
	}

	return serv
}
