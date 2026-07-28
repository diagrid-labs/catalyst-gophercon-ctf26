package flags

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileBonus(t *testing.T) {
	dir := t.TempDir()

	withBonus := filepath.Join(dir, "flags.json")
	if err := os.WriteFile(withBonus, []byte(`{
		"challenges": [{"id": 1, "tool": "board_tour", "prevCode": "c", "flag": "gc26{f}"}],
		"bonus": {"tool": "nedrys_backdoor", "flag": "gc26{bonus}"}
	}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadFile(withBonus)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	b, ok := s.Bonus()
	if !ok {
		t.Fatal("expected a bonus flag")
	}
	if b.Tool != "nedrys_backdoor" || b.Flag != "gc26{bonus}" {
		t.Errorf("bonus = %+v", b)
	}

	noBonus := filepath.Join(dir, "nobonus.json")
	if err := os.WriteFile(noBonus, []byte(`{"challenges":[{"id":1,"tool":"t","prevCode":"c","flag":"f"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s2, err := LoadFile(noBonus)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if _, ok := s2.Bonus(); ok {
		t.Error("expected no bonus flag")
	}
}
