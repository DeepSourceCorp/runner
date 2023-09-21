package testutil

type MockReader struct {
	Err     error
	Payload []byte
}

func (r *MockReader) Read(p []byte) (n int, err error) {
	if r.Err != nil {
		return 0, r.Err
	}
	return copy(p, r.Payload), nil
}
