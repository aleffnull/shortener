package config

type MemoryStoreConfiguration struct {
	KeyLength        int `validate:"required,ltefield=KeyMaxLength"`
	KeyMaxLength     int `validate:"required"`
	KeyMaxIterations int `validate:"required"`
}

func defaultMemoryStoreConfiguration() *MemoryStoreConfiguration {
	return &MemoryStoreConfiguration{
		KeyLength:        8,
		KeyMaxLength:     100,
		KeyMaxIterations: 10,
	}
}
