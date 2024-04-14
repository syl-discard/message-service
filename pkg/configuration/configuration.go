package configuration

type Configuration struct {
	DatabaseSettings DatabaseSettings
	APISettings      APISettings
}

type DatabaseSettings struct {
	Url      string
	Keyspace string
}

type APISettings struct {
	Address string
	Port    string
}
