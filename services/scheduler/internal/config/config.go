package config

// Config holds the scheduler configuration.
type Config struct {
	API    API    `mapstructure:"api"`
	Worker Worker `mapstructure:"worker"`
}

// API holds the API connection settings.
type API struct {
	URL string `mapstructure:"url"`
}

// Worker holds the worker authentication settings.
type Worker struct {
	Secret string `mapstructure:"secret"`
}
