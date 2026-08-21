package config

type Config struct {
	Values      map[string]string
	Application ApplicationConfig
	Database    DatabaseConfig
}

func (c *Config) String(key string) string {
	return c.Values[key]
}
