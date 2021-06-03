package metrics

// Config contains the configuration for the metric collection.
type Config struct {
	Enabled          bool   `toml:",omitempty"`
	EnabledExpensive bool   `toml:",omitempty"`
	HTTP             string `toml:",omitempty"`
	Port             int    `toml:",omitempty"`
	EnableInfluxDB   bool   `toml:",omitempty"`
	InfluxDBEndpoint string `toml:",omitempty"`
	InfluxDBDatabase string `toml:",omitempty"`
	InfluxDBUsername string `toml:",omitempty"`
	InfluxDBPassword string `toml:",omitempty"`
	InfluxDBTags     string `toml:",omitempty"`
}

// DefaultConfig is the default config for metrics used in go-ethereum.
var DefaultConfig = Config{
	Enabled:          false,
	EnabledExpensive: false,
	HTTP:             "127.0.0.1",
	Port:             6060,
	EnableInfluxDB:   false,
	InfluxDBEndpoint: "http://localhost:8086",
	InfluxDBDatabase: "fullnode",
	InfluxDBUsername: "test",
	InfluxDBPassword: "test",
	InfluxDBTags:     "host=localhost",
}
