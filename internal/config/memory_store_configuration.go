package config

import "fmt"

type MemoryStoreConfiguration struct {
	KeyLength        int `validate:"required,ltefield=KeyMaxLength"`
	KeyMaxLength     int `validate:"required"`
	KeyMaxIterations int `validate:"required"`
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
		KeyLength:        8,
		KeyMaxLength:     100,
		KeyMaxIterations: 10,
	}
}
