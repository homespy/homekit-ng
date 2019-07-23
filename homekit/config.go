package homekit

import (
	"io/ioutil"

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"

	"homekit-ng/homekit/device"
	"homekit-ng/homekit/publish"
)

type Config struct {
	Logging  LoggingConfig
	Tracking TrackingConfig
	Broker   BrokerConfig
	Influx   publish.InfluxConfig
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
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
