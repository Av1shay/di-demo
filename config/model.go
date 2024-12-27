package config

type DataSource string

const (
	DataSourceMySQL DataSource = "mysql"
	DataSourceMongo DataSource = "mongo"
)

var allDataSources = []DataSource{DataSourceMySQL, DataSourceMongo}

type CacheProvider string

const (
	CacheProviderRedis CacheProvider = "redis"
)

var allCacheProviders = []CacheProvider{CacheProviderRedis}
