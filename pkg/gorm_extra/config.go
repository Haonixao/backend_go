package gorm_extra

type Postgres struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Db       string `mapstructure:"db"`
	Port     string `mapstructure:"port"`
	Sslmode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"time_zone"`
}

var PostgresDefaults = map[string]any{
	"postgres.host":      "localhost",
	"postgres.port":      "5432",
	"postgres.user":      "main",
	"postgres.password":  "password",
	"postgres.db":        "main",
	"postgres.sslmode":   "disable",
	"postgres.time_zone": "Europe/Moscow",
}
