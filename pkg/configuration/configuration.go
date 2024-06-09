package configuration

type Configuration struct {
	DatabaseSettings DatabaseSettings
	APISettings      APISettings
}

type DatabaseSettings struct {
	Url        string
	Keyspace   string
	Provider   string
	AstraId    string // only used when provider is astra
	AstraToken string // only used when provider is astra
}

type APISettings struct {
	Address string
	Port    string
}
