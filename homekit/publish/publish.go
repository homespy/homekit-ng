package publish

import (
	"context"
	"fmt"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"go.uber.org/zap"

	"homekit-ng/homekit/tm"
)

type InfluxConfig struct {
	Addr     string
	Username string
	Password string
	Interval time.Duration
}

type InfluxDBMetricsWriter struct {
	cfg       *InfluxConfig
	telemetry *tm.TelemetryStorage
	log       *zap.SugaredLogger
}

func NewInfluxDBMetricsWriter(cfg *InfluxConfig, telemetry *tm.TelemetryStorage, log *zap.SugaredLogger) *InfluxDBMetricsWriter {
	return &InfluxDBMetricsWriter{
		cfg:       cfg,
		telemetry: telemetry,
		log:       log,
	}
}

func (m *InfluxDBMetricsWriter) Run(ctx context.Context) error {
	influx, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     m.cfg.Addr,
		Username: m.cfg.Username,
		Password: m.cfg.Password,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err := influx.Close(); err != nil {
			m.log.Warnw("failed to close InfluxDB client", zap.Error(err))
		}
	}()

	timer := time.NewTicker(m.cfg.Interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := m.push(ctx, influx); err != nil {
				m.log.Warnw("failed to push telemetry", zap.Error(err))
			}
		}
	}
}

func (m *InfluxDBMetricsWriter) push(ctx context.Context, influx influxdb.Client) error {
	fields := map[string]interface{}{}
	for _, telemetry := range m.telemetry.Read("/") {
		fields[telemetry.Topic] = telemetry.Value
	}

	point, err := influxdb.NewPoint("metrics", map[string]string{}, fields, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create InfluxDB point: %v", err)
	}

	pointsConfig := influxdb.BatchPointsConfig{
		Precision:        "s",
		Database:         "homekit",
		RetentionPolicy:  "",
		WriteConsistency: "all",
	}
	points, err := influxdb.NewBatchPoints(pointsConfig)
	if err != nil {
		return fmt.Errorf("failed to create InfluxDB telemetry: %v", err)
	}
	points.AddPoint(point)

	if err := influx.Write(points); err != nil {
		return fmt.Errorf("failed to write InfluxDB points: %v", err)
	}

	m.log.Debugf("pushed metrics to InfluxDB")

	return nil
}
