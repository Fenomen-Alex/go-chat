package config

type Config struct {
	Identity  IdentityConfig  `yaml:"identity" json:"identity" toml:"identity"`
	Network   NetworkConfig   `yaml:"network" json:"network" toml:"network"`
	Database  DatabaseConfig  `yaml:"database" json:"database" toml:"database"`
	Downloads DownloadsConfig `yaml:"downloads" json:"downloads" toml:"downloads"`
	Uploads   UploadsConfig   `yaml:"uploads" json:"uploads" toml:"uploads"`
	Theme     string          `yaml:"theme" json:"theme" toml:"theme"`
	Notify    NotifyConfig    `yaml:"notifications" json:"notifications" toml:"notifications"`
	Logging   LoggingConfig   `yaml:"logging" json:"logging" toml:"logging"`
	Security  SecurityConfig  `yaml:"security" json:"security" toml:"security"`
}

type IdentityConfig struct {
	DisplayName string `yaml:"display_name" json:"display_name" toml:"display_name"`
}

type NetworkConfig struct {
	Port           int      `yaml:"port" json:"port" toml:"port"`
	BootstrapPeers []string `yaml:"bootstrap_peers" json:"bootstrap_peers" toml:"bootstrap_peers"`
	RelayPeers     []string `yaml:"relay_peers" json:"relay_peers" toml:"relay_peers"`
	EnableRelay    bool     `yaml:"enable_relay" json:"enable_relay" toml:"enable_relay"`
	EnableMDNS     bool     `yaml:"enable_mdns" json:"enable_mdns" toml:"enable_mdns"`
	EnableDHT      bool     `yaml:"enable_dht" json:"enable_dht" toml:"enable_dht"`
	EnableQUIC     bool     `yaml:"enable_quic" json:"enable_quic" toml:"enable_quic"`
	EnableTCP      bool     `yaml:"enable_tcp" json:"enable_tcp" toml:"enable_tcp"`
}

type DatabaseConfig struct {
	Path    string `yaml:"path" json:"path" toml:"path"`
	Encrypt bool   `yaml:"encrypt" json:"encrypt" toml:"encrypt"`
}

type DownloadsConfig struct {
	Path    string `yaml:"path" json:"path" toml:"path"`
	MaxSize int64  `yaml:"max_size" json:"max_size" toml:"max_size"`
}

type UploadsConfig struct {
	MaxSize int64 `yaml:"max_size" json:"max_size" toml:"max_size"`
}

type NotifyConfig struct {
	Desktop  bool `yaml:"desktop" json:"desktop" toml:"desktop"`
	Bell     bool `yaml:"bell" json:"bell" toml:"bell"`
	Mentions bool `yaml:"mentions" json:"mentions" toml:"mentions"`
}

type LoggingConfig struct {
	Level  string `yaml:"level" json:"level" toml:"level"`
	File   string `yaml:"file" json:"file" toml:"file"`
	Rotate bool   `yaml:"rotate" json:"rotate" toml:"rotate"`
}

type SecurityConfig struct {
	KeyRotationDays int  `yaml:"key_rotation_days" json:"key_rotation_days" toml:"key_rotation_days"`
	EncryptDatabase bool `yaml:"encrypt_database" json:"encrypt_database" toml:"encrypt_database"`
}

func Default() *Config {
	return &Config{
		Identity: IdentityConfig{
			DisplayName: "",
		},
		Network: NetworkConfig{
			Port:        0,
			EnableRelay: true,
			EnableMDNS:  true,
			EnableDHT:   false,
			EnableQUIC:  true,
			EnableTCP:   true,
		},
		Database: DatabaseConfig{
			Path:    "chat.db",
			Encrypt: false,
		},
		Downloads: DownloadsConfig{
			Path:    "downloads",
			MaxSize: 100 * 1024 * 1024,
		},
		Uploads: UploadsConfig{
			MaxSize: 50 * 1024 * 1024,
		},
		Theme: "dark",
		Notify: NotifyConfig{
			Desktop:  false,
			Bell:     true,
			Mentions: true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			File:   "",
			Rotate: true,
		},
		Security: SecurityConfig{
			KeyRotationDays: 30,
			EncryptDatabase: false,
		},
	}
}
