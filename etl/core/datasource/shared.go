package datasource

import (
	"database/sql"
	"fmt"
	"sync"
)

// Shared wraps a datasource instance and exposes ref-counted leases so multiple
// components can safely share one initialized datasource without prematurely
// closing the underlying resource.
type Shared struct {
	base   Datasource
	mu     sync.Mutex
	refs   int
	closed bool
}

type sharedLease struct {
	shared *Shared
	once   sync.Once
}

func NewShared(base Datasource) (*Shared, error) {
	if base == nil {
		return nil, fmt.Errorf("datasource is nil")
	}
	return &Shared{base: base}, nil
}

func (s *Shared) Lease() Datasource {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return &sharedLease{}
	}
	s.refs++
	return &sharedLease{shared: s}
}

func (s *Shared) release() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	if s.refs > 0 {
		s.refs--
	}
	if s.refs != 0 {
		return nil
	}
	s.closed = true
	return s.base.Close()
}

func (l *sharedLease) Init(config map[string]string) error {
	if l.shared == nil || l.shared.base == nil {
		return fmt.Errorf("datasource lease is closed")
	}
	return l.shared.base.Init(config)
}

func (l *sharedLease) Close() error {
	if l.shared == nil {
		return nil
	}
	var err error
	l.once.Do(func() {
		err = l.shared.release()
	})
	return err
}

func (l *sharedLease) DB() *sql.DB {
	if l.shared == nil || l.shared.base == nil {
		return nil
	}
	return l.shared.base.DB()
}

func (l *sharedLease) ConfigMap() map[string]string {
	if l.shared == nil || l.shared.base == nil {
		return nil
	}
	return l.shared.base.ConfigMap()
}

func (l *sharedLease) ListTables() ([]TableInfo, error) {
	if l.shared == nil || l.shared.base == nil {
		return []TableInfo{}, nil
	}
	return l.shared.base.ListTables()
}
