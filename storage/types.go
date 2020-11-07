package storage

type Store interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Close() error
}
