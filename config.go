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
	// If true the linter will treat the migration as running inside a transaction by default.
	ImplicitTransaction *bool                     `yaml:"implicitTransaction"`
	Rules               map[rules.Code]ruleConfig `yaml:"rules"`
}

func defaultConfig() Config {
	ruleConfigs := map[rules.Code]ruleConfig{}
	for _, rule := range rules.AllRules() {
		enabled := !rule.DisabledByDefault
		ruleConfigs[rule.Code] = ruleConfig{Enabled: enabled}
	}

	implicitTx := true
	return Config{
		ImplicitTransaction: &implicitTx,
		Rules:               ruleConfigs,
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

	if parsed.ImplicitTransaction != nil {
		cfg.ImplicitTransaction = parsed.ImplicitTransaction
	}

	return cfg, nil
}
