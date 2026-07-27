package flags

import "crypto/subtle"

// Resolve returns the flag for challengeID when submitted matches that
// challenge's required access code, compared in constant time. The boolean
// reports whether the unlock succeeded; the returned flag is empty on failure,
// and the correct code is never revealed to the caller.
func (s *Store) Resolve(challengeID int, submitted string) (string, bool) {
	c, ok := s.byID[challengeID]
	if !ok {
		return "", false
	}
	if subtle.ConstantTimeCompare([]byte(submitted), []byte(c.PrevCode)) == 1 {
		return c.Flag, true
	}
	return "", false
}
