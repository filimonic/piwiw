package piwiw

import (
	"encoding/json"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ProxyPort                 int    `env:"SERVER_PORT,PROXY_PORT" env-default:"11434"`
	OpenAIAPIBaseUrl          string `env:"OPENAI_API_BASE_URL,required"`
	SkipTLSVerify             bool   `env:"SKIP_TLS_VERIFY" env-default:"false"`
	OpenAIAPIToken            string `env:"OPENAI_API_KEY,required"`
	OpenAIAPIChatForcedParams string `env:"OPENAI_API_CHAT_FORCED_PARAMS"`
	RequestTimeout            int    `env:"REQUEST_TIMEOUT" env-default:"180"`
	MaxRetries                int    `env:"MAX_RETRIES" env-default:"3"`
	RetryDelay                int    `env:"RETRY_DELAY" env-default:"300"`
	EmptyContentText          string `env:"EMPTY_CONTENT_TEXT" env-default:""`
	TraceFolderPath           string `env:"TRACE_FOLDER_PATH"`
	TraceKeepHours            int    `env:"TRACE_KEEP_HOURS" env-default:"2160"`
}

func (c *Config) GetOpenAIAPIChatForcedParams() json.RawMessage {
	if c.OpenAIAPIChatForcedParams == "" {
		return nil
	}
	return json.RawMessage(c.OpenAIAPIChatForcedParams)
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	if cfg.ProxyPort <= 0 || cfg.ProxyPort >= 65535 {
		return nil, fmt.Errorf("PORT is not valid")
	}
	if cfg.OpenAIAPIChatForcedParams != "" && !json.Valid([]byte(cfg.OpenAIAPIChatForcedParams)) {
		return nil, fmt.Errorf("OPENAI_API_CHAT_FORCED_PARAMS is not valid JSON")
	}
	if cfg.RequestTimeout <= 0 {
		return nil, fmt.Errorf("REQUEST_TIMEOUT must be greater than 0")
	}
	if cfg.MaxRetries < 0 {
		return nil, fmt.Errorf("MAX_RETRIES must be greater or equal to 0")
	}
	if cfg.RetryDelay < 0 {
		cfg.RetryDelay = 0
	}
	if cfg.TraceKeepHours <= 0 {
		return nil, fmt.Errorf("TRACE_KEEP_HOURS must be greater than 0")
	}
	return &cfg, nil
}

func SetConfig(config *Config) {
	cfg = config
}
