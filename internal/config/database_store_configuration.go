package config

import "fmt"

type DatabaseStoreConfiguration struct {
	KeyStoreConfiguration
	DataSourceName string `env:"DATABASE_DSN"`
}

func (c *DatabaseStoreConfiguration) String() string {
	return fmt.Sprintf(
		"&DatabaseStoreConfiguration{KeyLength:%v KeyMaxLength:%v KeyMaxIterations:%v DataSourceName:'%v'}",
		c.KeyLength,
		c.KeyMaxLength,
		c.KeyMaxIterations,
		c.DataSourceName,
	)
}

func NewDatabaseStoreConfiguration(dataSourceName string) *DatabaseStoreConfiguration {
	return &DatabaseStoreConfiguration{
		KeyStoreConfiguration: KeyStoreConfiguration{
			KeyLength:        8,
			KeyMaxLength:     100,
			KeyMaxIterations: 3,
		},
		DataSourceName: dataSourceName,
	}
}

func (c *DatabaseStoreConfiguration) IsDatabaseEnabled() bool {
	return len(c.DataSourceName) != 0
}
