package config

type ConfigurationFile struct {
	ServerAddress               string `json:"server_address"`
	ServerAddressGRPC           string `json:"server_address_grpc"`
	BaseURL                     string `json:"base_url"`
	AuditFile                   string `json:"audit_file"`
	AuditURL                    string `json:"audit_url"`
	FileStoreFilePath           string `json:"file_storage_path"`
	DatabaseStoreDataSourceName string `json:"database_dsn"`
	HTTPSEnabled                bool   `json:"enable_https"`
	HTTPSCertificateFile        string `json:"https_certificate_file"`
	HTTPSKeyFile                string `json:"https_key_file"`
	CPUProfile                  string `json:"cpu_profile"`
	MemoryProfile               string `json:"memory_profile"`
	TrustedSubnet               string `json:"trusted_subnet"`
}
