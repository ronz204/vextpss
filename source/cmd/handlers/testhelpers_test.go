package handlers_test

// passwordResponse bundles a ReadPassword return value.
type passwordResponse struct {
	bytes []byte
	err   error
}

// sequentialPrompter returns pre-set responses in order for each call type.
// Useful when a handler calls ReadLine/ReadPassword multiple times with different expected outcomes.
type sequentialPrompter struct {
	lineResponses     []string
	lineErrors        []error
	passwordResponses []passwordResponse
	confirmResponse   bool
	confirmErr        error
	lineIdx           int
	passwordIdx       int
}

func (s *sequentialPrompter) ReadLine(_ string) (string, error) {
	if s.lineIdx >= len(s.lineResponses) {
		return "", nil
	}
	resp := s.lineResponses[s.lineIdx]
	var err error
	if s.lineIdx < len(s.lineErrors) {
		err = s.lineErrors[s.lineIdx]
	}
	s.lineIdx++
	return resp, err
}

func (s *sequentialPrompter) ReadPassword(_ string) ([]byte, error) {
	if s.passwordIdx >= len(s.passwordResponses) {
		return []byte("default"), nil
	}
	resp := s.passwordResponses[s.passwordIdx]
	s.passwordIdx++
	return resp.bytes, resp.err
}

func (s *sequentialPrompter) Confirm(_ string) (bool, error) {
	return s.confirmResponse, s.confirmErr
}

func (s *sequentialPrompter) Zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
