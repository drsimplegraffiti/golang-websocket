package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

/*
structs are used to group related configuration parameters together, making
it easier to manage and access them. In this case, we have a struct for the HTTP
server configuration and a main Config struct that includes the HTTP server
configuration along with other application settings.
*/
type HTTPServer struct {
	Address string `env:"HTTP_ADDRESS" env-default: "localhost:8082"`
}

type Config struct {
	ENV    string `env:"ENV" env-default: "dev"`
	DBPath string `env:"DB_PATH" env-default: "sqlite/dev"`
	DBName string `env:"DB_NAME" env-default: "api.db"`
	HTTPServer
	JWTKey string `env:"JWT_KEY" env-default: "supersecretkey"`
}

// LoadConfig is like a constructor for the Config struct. It reads the
// configuration from the specified .env file or environment variables and populates the Config
func LoadConfig() *Config {
	var cfg Config
	var envPath string

	/*
	   flag package provides a way to define and parse command-line flags. In this
	   case, we are defining a flag named "config" that allows the user to specify
	   the path to a .env file. The value of this flag will be stored in the
	   envPath variable.
	*/
	flag.StringVar(&envPath, "config", "", "path to .env file")
	flag.Parse()

	if envPath == "" {
		envPath = os.Getenv("CONFIG_PATH")
	}

	if envPath == "" {
		envPath = "config/dev.env"
	}

	err := cleanenv.ReadConfig(envPath, &cfg)
	if err != nil {
		log.Fatalf("cannot read .env from %s, %v", envPath, err)
	}

	return &cfg
}
