package storage

import (
	"context"
	"fmt"
	"io"
	"math"
	"regexp"
	"sync"
)

func NewManager() *Manager {
	return &Manager{
		mu: &sync.RWMutex{},
	}
}

var _ Interface = (*Manager)(nil)

type Manager struct {
	matchers []func(mime string, size int64) bool
	drivers  []Interface

	mu *sync.RWMutex
}

func (m *Manager) Name() string {
	return ""
}

func (m *Manager) Add(mimeMatch string, maxSize int64, config Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		exp *regexp.Regexp
		err error
	)
	if len(mimeMatch) != 0 {
		exp, err = regexp.CompilePOSIX(mimeMatch)
		if err != nil {
			return fmt.Errorf("invalid mime match regex: %w", err)
		}
	}

	d, err := config.Create()
	if err != nil {
		return fmt.Errorf("failed to create driver from config %T: %w", config, err)
	}

	m.drivers = append(m.drivers, d)

	if maxSize == 0 {
		maxSize = math.MaxInt64
	}

	if exp != nil {
		m.matchers = append(m.matchers, func(mime string, size int64) bool {
			return exp.MatchString(mime) && (size < maxSize)
		})
	} else {
		m.matchers = append(m.matchers, func(mime string, size int64) bool {
			return size < maxSize
		})
	}

	return nil
}

func (m *Manager) Upload(
	ctx context.Context, filename, mimeType string, size int64, data io.Reader,
) (url string, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.drivers) == 0 {
		// no driver, storage disabled
		return "", nil
	}

	for i, match := range m.matchers {
		if !match(mimeType, size) {
			continue
		}

		return m.drivers[i].Upload(ctx, filename, mimeType, size, data)
	}

	return "", fmt.Errorf("not handled by any storage driver")
}
