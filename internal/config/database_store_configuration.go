package config

import "fmt"

type DatabaseStoreConfiguration struct {
	DataSourceName string `env:"DATABASE_DSN"`
}

func (c *DatabaseStoreConfiguration) String() string {
	return fmt.Sprintf("&DatabaseStoreConfiguration{DataSourceName:'%v'}", c.DataSourceName)
}

func (c *DatabaseStoreConfiguration) IsDatabaseEnabled() bool {
	return len(c.DataSourceName) != 0
}
