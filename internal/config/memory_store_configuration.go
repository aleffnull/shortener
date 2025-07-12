package config

import "fmt"

type MemoryStoreConfiguration struct {
	KeyStoreConfiguration
}

func (c *MemoryStoreConfiguration) String() string {
	return fmt.Sprintf(
		"&MemoryStoreConfiguration{KeyLength:%v KeyMaxLength:%v KeyMaxIterations:%v}",
		c.KeyLength,
		c.KeyMaxLength,
		c.KeyMaxIterations)
}

func defaultMemoryStoreConfiguration() *MemoryStoreConfiguration {
	return &MemoryStoreConfiguration{
		KeyStoreConfiguration: KeyStoreConfiguration{
			KeyLength:        8,
			KeyMaxLength:     100,
			KeyMaxIterations: 10,
		},
	}
}
