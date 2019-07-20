package homekit

import (
	"encoding/json"
	"io/ioutil"

	"go.uber.org/zap/zapcore"

	"homekit-ng/homekit/device"
)

type Config struct {
	Logging  LoggingConfig
	Tracking TrackingConfig
	Broker   BrokerConfig
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

type LoggingConfig struct {
	Level zapcore.Level
}

type TrackingConfig struct {
	Devices map[string]*device.TrackingConfig
}

type BrokerConfig struct {
	Port uint16
}
