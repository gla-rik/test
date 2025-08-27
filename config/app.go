package config

type App struct {
	Host string `envconfig:"APP_HOST" default:"localhost"`
	Port string `envconfig:"APP_PORT" default:"8080"`
}
