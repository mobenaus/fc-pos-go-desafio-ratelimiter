package configs

import "github.com/spf13/viper"

type Conf struct {
	RateLimitStrategy string `mapstructure:"RATE_LIMIT_STRATEGY"`
	REDISAddr         string `mapstructure:"REDIS_ADDR"`
	REDISPassword     string `mapstructure:"REDIS_PASSWORD"`
	REDISDefaultDB    int    `mapstructure:"REDIS_DEFAULT_DB"`
	IPRateLimit       int    `mapstructure:"IP_RATE_LIMIT"`
	IPRatePeriod      string `mapstructure:"IP_RATE_PERIOD"`
	TOKENRateLimit    int    `mapstructure:"TOKEN_RATE_LIMIT"`
	TOKENRatePeriod   string `mapstructure:"TOKEN_RATE_PERIOD"`
	WebServerPort     string `mapstructure:"WEB_SERVER_PORT"`
}

func LoadConfig(path string) (*Conf, error) {
	var cfg *Conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg, err
}
