// Package flags loads the per-challenge access codes and flag values and
// resolves unlock attempts. Real values are never committed: they are loaded at
// boot from a JSON file referenced by the environment — a mounted secret in
// prod, a fake fixture in dev.
package flags

import (
	"encoding/json"
	"fmt"
	"os"
)

// Challenge is one link in the unlock chain: submitting PrevCode to Tool yields Flag.
// For Challenge 1, PrevCode is the bootstrap ticket published in the handout; for
// every later challenge it is the previous challenge's flag.
type Challenge struct {
	ID       int    `json:"id"`
	Tool     string `json:"tool"`
	PrevCode string `json:"prevCode"`
	Flag     string `json:"flag"`
}

// BonusFlag is an optional hidden flag returned by an undocumented tool, outside
// the main unlock chain — discovering the tool is the whole challenge.
type BonusFlag struct {
	Tool string `json:"tool"`
	Flag string `json:"flag"`
}

// config is the on-disk shape of the flags file.
type config struct {
	Challenges []Challenge `json:"challenges"`
	Bonus      *BonusFlag  `json:"bonus,omitempty"`
}

// Store holds the challenge chain in memory, keyed by challenge ID, plus an
// optional hidden bonus flag.
type Store struct {
	byID  map[int]Challenge
	bonus *BonusFlag
}

// LoadFile reads and validates a flags JSON file into a Store.
func LoadFile(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read flags file %q: %w", path, err)
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse flags file %q: %w", path, err)
	}
	s, err := New(cfg.Challenges)
	if err != nil {
		return nil, err
	}
	if cfg.Bonus != nil {
		if cfg.Bonus.Tool == "" || cfg.Bonus.Flag == "" {
			return nil, fmt.Errorf("bonus flag missing tool or flag")
		}
		s.bonus = cfg.Bonus
	}
	return s, nil
}

// New builds a Store from challenges, validating each entry and rejecting
// duplicates. It returns an error if any field is missing so a misconfigured
// deploy fails fast at boot rather than silently withholding flags.
func New(challenges []Challenge) (*Store, error) {
	if len(challenges) == 0 {
		return nil, fmt.Errorf("no challenges configured")
	}
	byID := make(map[int]Challenge, len(challenges))
	for _, c := range challenges {
		switch {
		case c.ID == 0:
			return nil, fmt.Errorf("challenge missing id")
		case c.Tool == "":
			return nil, fmt.Errorf("challenge %d missing tool", c.ID)
		case c.PrevCode == "":
			return nil, fmt.Errorf("challenge %d missing prevCode", c.ID)
		case c.Flag == "":
			return nil, fmt.Errorf("challenge %d missing flag", c.ID)
		}
		if _, dup := byID[c.ID]; dup {
			return nil, fmt.Errorf("duplicate challenge id %d", c.ID)
		}
		byID[c.ID] = c
	}
	return &Store{byID: byID}, nil
}

// Bonus returns the configured hidden bonus flag, if any.
func (s *Store) Bonus() (BonusFlag, bool) {
	if s.bonus == nil {
		return BonusFlag{}, false
	}
	return *s.bonus, true
}
