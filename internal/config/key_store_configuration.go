package config

type KeyStoreConfiguration struct {
	KeyLength        int `validate:"required,ltefield=KeyMaxLength"`
	KeyMaxLength     int `validate:"required"`
	KeyMaxIterations int `validate:"required"`
}
