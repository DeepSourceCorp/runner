package session

type Store interface {
	Create(*Session) error
	Update(*Session) error
	Get(id string) (*Session, error)
	GetByCode(code string) (*Session, error)
}
