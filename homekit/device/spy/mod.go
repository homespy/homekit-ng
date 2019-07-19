package spy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"homekit-ng/homekit/device"
)

type activityData struct {
	MAC        net.HardwareAddr
	IP         net.IP
	LastSeen   time.Time
	CancelFunc context.CancelFunc
}

type Spy struct {
	IdleTimeout time.Duration
	device      string
	// The reason we do everything through this channel is to share the main
	// context provided in "Run" method.
	txrx        chan interface{}
	mu          sync.RWMutex
	activity    map[string]*activityData
	log         *zap.SugaredLogger
}

func NewSpy(device string, log *zap.SugaredLogger) *Spy {
	txrx := make(chan interface{}, 128)

	return &Spy{
		IdleTimeout: 5 * time.Minute,
		device:      device,
		txrx:        txrx,
		activity:    map[string]*activityData{},
		log:         log,
	}
}

func (m *Spy) IsUp(mac net.HardwareAddr) bool {
	return time.Now().Sub(m.HardwareLastSeen(mac)) < m.IdleTimeout
}

func (m *Spy) HardwareLastSeen(mac net.HardwareAddr) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.activity[mac.String()]
	if ok {
		return info.LastSeen
	}

	return time.Time{}
}

func (m *Spy) Register(mac net.HardwareAddr) {
	m.txrx <- &registerEvent{MAC: mac}
}

func (m *Spy) Run(ctx context.Context) error {
	m.updateHardwareCache(ctx)

	timer := time.NewTicker(10 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			m.updateHardwareCache(ctx)
		case event := <-m.txrx:
			switch ev := event.(type) {
			case *registerEvent:
				m.watchDevice(ctx, ev.MAC)
			default:
				return fmt.Errorf("unknown event type: %T", ev)
			}
		}
	}
}

func (m *Spy) watchDevice(ctx context.Context, mac net.HardwareAddr) {
	ctx, cancelFunc := context.WithCancel(ctx)

	m.mu.Lock()
	defer m.mu.Unlock()

	activity, ok := m.activity[mac.String()]
	if ok {
		activity.CancelFunc()
	}

	m.activity[mac.String()] = &activityData{
		MAC:        mac,
		IP:         nil,
		LastSeen:   time.Time{},
		CancelFunc: cancelFunc,
	}
}

func (m *Spy) updateHardwareCache(ctx context.Context) {
	m.log.Debug("updating ARP cache")
	defer m.log.Debug("updated ARP cache")

	scanner := device.NewScanner(m.device)
	scanInfo, err := scanner.Scan(ctx)
	if err != nil {
		m.log.Warnw("failed to scan for devices", zap.Error(err))
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, record := range scanInfo {
		activity, ok := m.activity[record.MAC]
		if ok {
			// Do nothing if the IP address hasn't changed.
			if activity.IP.Equal(record.IP) {
				continue
			}

			m.log.Infof("updated IP address for %s: %s -> %s", activity.MAC, activity.IP, record.IP)
			activity.CancelFunc()

			ctx, cancelFunc := context.WithCancel(ctx)
			activity.IP = record.IP
			activity.LastSeen = time.Time{}
			activity.CancelFunc = cancelFunc

			m.spawnWatcher(ctx, activity.MAC, activity.IP)
		}
	}
}

func (m *Spy) spawnWatcher(ctx context.Context, mac net.HardwareAddr, ip net.IP) {
	go func() {
		if err := m.watch(ctx, mac, ip); err != nil {
			m.log.Warnf("stopped watching for %s: %v", mac, err)
		}
	}()
}

func (m *Spy) watch(ctx context.Context, mac net.HardwareAddr, ip net.IP) error {
	m.log.Infof("watching for %s", mac)
	defer m.log.Infof("stopped watching for %s", mac)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		sc := &PingCheck{
			Addr:     ip.String(),
			Interval: 1 * time.Second,
			OnActivity: func() {
				m.log.Info("detected PING activity")
				m.updateLastSeen(mac)
			},
			Log: m.log,
		}

		return sc.Run(ctx)
	})
	wg.Go(func() error {
		sc := &SYNCheck{
			Addr:     fmt.Sprintf("%s:62078", ip.String()),
			Interval: 1 * time.Second,
			OnActivity: func() {
				m.log.Info("detected SYN activity")
				m.updateLastSeen(mac)
			},
			Log: m.log,
		}

		return sc.Run(ctx)
	})
	wg.Go(func() error {
		filter, err := CompileFilter(mac)
		if err != nil {
			return err
		}

		sc := &PCapCheck{
			Device: m.device,
			Filter: filter,
			OnActivity: func() {
				m.log.Info("detected PCAP activity")
				m.updateLastSeen(mac)
			},
			Log: m.log,
		}

		return sc.Run(ctx)
	})

	<-ctx.Done()

	return wg.Wait()
}

func (m *Spy) updateLastSeen(mac net.HardwareAddr) {
	m.mu.Lock()
	defer m.mu.Unlock()

	activity, ok := m.activity[mac.String()]
	if ok {
		activity.LastSeen = time.Now()
	}
}

type registerEvent struct {
	MAC net.HardwareAddr
}
