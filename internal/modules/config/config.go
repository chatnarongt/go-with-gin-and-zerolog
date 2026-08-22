package config

type Config struct {
	Values      map[string]string
	Application ApplicationConfig
	Databases   DatabasesConfig
}

func (c *Config) String(key string) string {
	return c.Values[key]
}
