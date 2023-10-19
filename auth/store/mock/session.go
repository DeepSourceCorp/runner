package mock

import (
	"errors"

	"github.com/deepsourcecorp/runner/auth/session"
)

type MockSessionStore struct {
	sessions []*session.Session
}

func NewInMemorySessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: []*session.Session{},
	}
}

func (s *MockSessionStore) Create(session *session.Session) error {
	s.sessions = append(s.sessions, session)
	return nil
}

func (store *MockSessionStore) Update(session *session.Session) error {
	for i, s := range store.sessions {
		if s.ID == session.ID {
			store.sessions[i] = session
			return nil
		}
	}
	return errors.New("session not found")
}

func (store *MockSessionStore) Get(id string) (*session.Session, error) {
	for _, s := range store.sessions {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, errors.New("session not found")
}

func (store *MockSessionStore) GetByCode(code string) (*session.Session, error) {
	for _, s := range store.sessions {
		if s.Code == code {
			return s, nil
		}
	}
	return nil, errors.New("session not found")
}
