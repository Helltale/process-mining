package config

import "time"

func (c *Config) GetAppMaxReadTime() time.Duration {
	return time.Duration(c.APP_MAX_READ_TIME) * time.Second
}

func (c *Config) GetAppMaxWriteTime() time.Duration {
	return time.Duration(c.APP_MAX_WRITE_TIME) * time.Second
}
