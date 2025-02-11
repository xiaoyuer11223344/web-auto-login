package config

import "time"

type SelectorConfig struct {
	UserInput     string `yaml:"userInput"`
	PasswordInput string `yaml:"passwordInput"`
	LoginBtn      string `yaml:"loginBtn"`
	CaptchaInput  string `yaml:"captchaInput"`
	CaptchaImg    string `yaml:"captchaImg"`
}

type Config struct {
	Inputs        []string
	InputsFile    string
	LogLevel      string
	OutputFile    string
	CrackAll      bool
	Delay         int
	Headless      bool
	MaxAttempts   int
	MaxCrackNum   int
	MaxCrackTime  int
	PassList      []string
	PassFile      string
	Proxy         string
	SelectorFile  string
	Threads       int
	UserList      []string
	UserFile      string
	LoginTimeout  time.Duration
	PageTimeout   time.Duration
}

func NewConfig() *Config {
	return &Config{
		LoginTimeout: DefaultLoginTimeout,
		PageTimeout:  DefaultPageTimeout,
	}
}
