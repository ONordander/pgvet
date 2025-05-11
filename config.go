package main

import (
	"os"

	"github.com/onordander/pgvet/rules"

	"github.com/goccy/go-yaml"
)

type ruleConfig struct {
	Enabled bool `yaml:"enabled"`
}

type Config struct {
	Rules map[rules.Code]ruleConfig `yaml:"rules"`
}

func defaultConfig() Config {
	ruleConfigs := map[rules.Code]ruleConfig{}
	for _, rule := range rules.AllRules() {
		enabled := !rule.DisabledByDefault
		ruleConfigs[rule.Code] = ruleConfig{Enabled: enabled}
	}

	return Config{
		Rules: ruleConfigs,
	}
}

func overlayConfig(cfg Config, path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var parsed Config
	if err := yaml.NewDecoder(f).Decode(&parsed); err != nil {
		return Config{}, err
	}

	for code, ruleConfig := range parsed.Rules {
		cfg.Rules[code] = ruleConfig
	}

	return cfg, nil
}
