package config

import (
	"errors"
	"fmt"
	"github.com/Av1shay/di-demo/uam"
	"os"
	"slices"
	"strconv"
)

type Manager struct {
	datasource    DataSource
	mysqlConn     string
	mongoURI      string
	mongoDB       string
	cacheProvider CacheProvider
	cacheEnabled  bool
	redisAddr     string
	redisPassword string
}

func NewManager() *Manager {
	m := &Manager{}
	m.datasource = DataSource(os.Getenv("DATA_SOURCE"))
	m.mysqlConn = os.Getenv("MYSQL_CONNECTION")
	m.mongoURI = os.Getenv("MONGO_URI")
	m.mongoDB = os.Getenv("MONGO_DB")
	m.cacheProvider = CacheProvider(os.Getenv("CACHE_PROVIDER"))
	m.cacheEnabled, _ = strconv.ParseBool(os.Getenv("CACHE_ENABLED"))
	m.redisAddr = os.Getenv("REDIS_ADDR")
	m.redisPassword = os.Getenv("REDIS_PASSWORD")
	return m
}

func (m *Manager) Validate() error {
	if !slices.Contains(allDataSources, m.datasource) {
		return fmt.Errorf("invalid data source provided %q, vlid values: %+v", m.datasource, allDataSources)
	}
	if m.datasource == DataSourceMySQL && m.mysqlConn == "" {
		return errors.New("mysql connection string is required")
	}
	if m.datasource == DataSourceMongo && (m.mongoURI == "" || m.mongoDB == "") {
		return errors.New("mongo uri and db name are required")
	}
	if !slices.Contains(allCacheProviders, m.cacheProvider) {
		return fmt.Errorf("invalid cache providor provided %q, vlid values: %+v", m.cacheProvider, allCacheProviders)
	}
	if m.cacheProvider == CacheProviderRedis && m.redisAddr == "" {
		return errors.New("redis address is required")
	}
	return nil
}

func (m *Manager) DataSource() DataSource {
	return m.datasource
}

func (m *Manager) MySQLConn() string {
	return m.mysqlConn
}

func (m *Manager) MongoURI() string {
	return m.mongoURI
}

func (m *Manager) MongoDB() string {
	return m.mongoDB
}

func (m *Manager) CacheEnabled() bool {
	return m.cacheEnabled
}

func (m *Manager) CacheProvider() CacheProvider {
	return m.cacheProvider
}

func (m *Manager) RedisAddr() string {
	return m.redisAddr
}

func (m *Manager) RedisPassword() string {
	return m.redisPassword
}

func (m *Manager) UAMAPIConfig() uam.Config {
	return uam.Config{
		CacheEnabled: m.CacheEnabled(),
	}
}
