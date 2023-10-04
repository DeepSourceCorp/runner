package session

import (
	"sync"
)

type MemoryStore struct {
	sessions []*Session
	lock     sync.RWMutex
}

func NewMockStore() Store {
	return &MemoryStore{
		sessions: []*Session{},
	}
}

func (m *MemoryStore) Create(session *Session) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.sessions = append(m.sessions, session)
	return nil
}

func (m *MemoryStore) Update(session *Session) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i, s := range m.sessions {
		if s.ID == session.ID {
			m.sessions[i] = session
			return nil
		}
	}
	return nil
}

func (m *MemoryStore) Filter(filter *Filter) (*Session, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, s := range m.sessions {
		if filter.ID != "" && filter.ID == s.ID {
			return s, nil
		}
		if filter.Code != "" && filter.Code == s.Code {
			return s, nil
		}
		if filter.RefreshToken != "" && filter.RefreshToken == s.RefreshToken {
			return s, nil
		}
	}
	return nil, nil
}
