package session

type Filter struct {
	ID           string
	Code         string
	RefreshToken string
}

type Store interface {
	Create(session *Session) error
	Update(session *Session) error
	Filter(filter *Filter) (*Session, error)
}
