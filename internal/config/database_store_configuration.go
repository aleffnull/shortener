package config

import (
	"fmt"

	"github.com/samber/lo"
)

type DatabaseStoreConfiguration struct {
	KeyStoreConfiguration
	DataSourceName string `env:"DATABASE_DSN"`
}

func (c *DatabaseStoreConfiguration) String() string {
	dataSourceName := lo.Ternary(len(c.DataSourceName) == 0, "", "*****")
	return fmt.Sprintf(
		"&DatabaseStoreConfiguration{KeyLength:%v KeyMaxLength:%v KeyMaxIterations:%v DataSourceName:'%v'}",
		c.KeyLength,
		c.KeyMaxLength,
		c.KeyMaxIterations,
		dataSourceName,
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
