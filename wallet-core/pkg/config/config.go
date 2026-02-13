package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	DB     DBConfig     `mapstructure:"db"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Wallet WalletConfig `mapstructure:"wallet"`
}

type AppConfig struct {
	Env      string `mapstructure:"env"`
	HttpPort string `mapstructure:"http_port"`
	GrpcPort string `mapstructure:"grpc_port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	MQType   string `mapstructure:"mq_type"` // "redis" or "kafka"
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
}

type WalletConfig struct {
	Mnemonic     string `mapstructure:"mnemonic"`
	HotWallet    string `mapstructure:"hot_wallet"`
	RpcUrl       string `mapstructure:"rpc_url"`
	KeystorePath string `mapstructure:"keystore_path"` // [NEW] 本地 Keystore 文件路径
	Password     string `mapstructure:"password"`      // [NEW] Keystore 密码 (通常通过环境变量 WALLET_PASSWORD 传入)
}

var Global Config

func Init() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.AddConfigPath("./config")

	// 环境变量设置
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Printf("Warning: Config file not found, using defaults and environment variables")
		} else {
			// Config file was found but another error was produced
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}

	if err := viper.Unmarshal(&Global); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	log.Printf("Configuration loaded successfully. Env: %s", Global.App.Env)
}

func setDefaults() {
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.http_port", "8080")
	viper.SetDefault("app.grpc_port", "50051")

	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("db.user", "wallet_user")
	viper.SetDefault("db.password", "wallet_password")
	viper.SetDefault("db.name", "wallet_db")

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.mq_type", "redis")

	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})

	viper.SetDefault("wallet.keystore_path", "wallet.json")
}
