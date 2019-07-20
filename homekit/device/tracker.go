package device

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type activityData struct {
	MAC        net.HardwareAddr
	IP         net.IP
	LastSeen   time.Time
	CancelFunc context.CancelFunc
}

type ActivityTracker struct {
	IdleTimeout time.Duration
	// The reason we do everything through this channel is to share the main
	// context provided in "Run" method.
	txrx     chan interface{}
	mu       sync.RWMutex
	activity map[string]*activityData
	log      *zap.SugaredLogger
}

func NewActivityTracker(log *zap.SugaredLogger) *ActivityTracker {
	txrx := make(chan interface{}, 128)

	return &ActivityTracker{
		IdleTimeout: 5 * time.Minute,
		txrx:        txrx,
		activity:    map[string]*activityData{},
		log:         log,
	}
}

func (m *ActivityTracker) IsUp(mac string) bool {
	return time.Now().Sub(m.HardwareLastSeen(mac)) < m.IdleTimeout
}

func (m *ActivityTracker) HardwareLastSeen(mac string) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.activity[mac]
	if ok {
		return info.LastSeen
	}

	return time.Time{}
}

func (m *ActivityTracker) Register(mac net.HardwareAddr, cfg *TrackingConfig) error {
	methods := make([]Tracker, len(cfg.Methods))
	for id, methodConfig := range cfg.Methods {
		factory, err := NewTrackingMethodFactory(methodConfig)
		if err != nil {
			return err
		}

		method, err := factory(mac, m.log)
		if err != nil {
			return err
		}

		methods[id] = method
	}

	m.txrx <- &registerEvent{MAC: mac, Trackers: methods}

	return nil
}

func (m *ActivityTracker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-m.txrx:
			switch ev := event.(type) {
			case *registerEvent:
				m.watchDevice(ctx, ev)
			default:
				return fmt.Errorf("unknown event type: %T", ev)
			}
		}
	}
}

func (m *ActivityTracker) watchDevice(ctx context.Context, ev *registerEvent) {
	ctx, cancelFunc := context.WithCancel(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	activity, ok := m.activity[ev.MAC.String()]
	if ok {
		activity.CancelFunc()
	}

	m.activity[ev.MAC.String()] = &activityData{
		MAC:        ev.MAC,
		IP:         nil,
		LastSeen:   time.Time{},
		CancelFunc: cancelFunc,
	}

	m.spawnWatcher(ctx, ev.MAC, ev.Trackers)
}

func (m *ActivityTracker) spawnWatcher(ctx context.Context, mac net.HardwareAddr, trackers []Tracker) {
	go func() {
		if err := m.watch(ctx, mac, trackers); err != nil {
			m.log.Warnf("stopped watching for %s: %v", mac, err)
		}
	}()
}

func (m *ActivityTracker) watch(ctx context.Context, mac net.HardwareAddr, trackers []Tracker) error {
	m.log.Infof("watching for %s", mac)
	defer m.log.Infof("stopped watching for %s", mac)

	wg, ctx := errgroup.WithContext(ctx)
	for _, tracker := range trackers {
		tracker := tracker

		wg.Go(func() error {
			return tracker.Run(ctx, func() {
				m.log.Infof("detected %T activity", tracker)
				m.updateLastSeen(mac)
			})
		})
	}

	<-ctx.Done()

	return wg.Wait()
}

func (m *ActivityTracker) updateLastSeen(mac net.HardwareAddr) {
	m.mu.Lock()
	defer m.mu.Unlock()

	activity, ok := m.activity[mac.String()]
	if ok {
		activity.LastSeen = time.Now()
	}
}

type registerEvent struct {
	MAC      net.HardwareAddr
	Trackers []Tracker
}
