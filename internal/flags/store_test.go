package flags

import "testing"

func TestNewValidation(t *testing.T) {
	tests := []struct {
		name       string
		challenges []Challenge
		wantErr    bool
	}{
		{"empty", nil, true},
		{"missing tool", []Challenge{{ID: 1, PrevCode: "c", Flag: "f"}}, true},
		{"missing prevCode", []Challenge{{ID: 1, Tool: "t", Flag: "f"}}, true},
		{"missing flag", []Challenge{{ID: 1, Tool: "t", PrevCode: "c"}}, true},
		{"missing id", []Challenge{{Tool: "t", PrevCode: "c", Flag: "f"}}, true},
		{"duplicate id", []Challenge{
			{ID: 1, Tool: "a", PrevCode: "c", Flag: "f"},
			{ID: 1, Tool: "b", PrevCode: "c", Flag: "g"},
		}, true},
		{"valid", []Challenge{{ID: 1, Tool: "t", PrevCode: "c", Flag: "f"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.challenges)
			if (err != nil) != tt.wantErr {
				t.Errorf("New err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFixtureChainIsConsistent loads the committed dev fixture and asserts the
// unlock chain is internally consistent: each challenge's prevCode equals the
// previous challenge's flag. This catches a broken chain in the fixture (and,
// by mirroring, in any real flags file authored to the same shape).
func TestFixtureChainIsConsistent(t *testing.T) {
	store, err := LoadFile("../../fixtures/flags.dev.json")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	// Walk 1..4: the flag returned for N must unlock N+1.
	prev := ""
	for id := 1; id <= 4; id++ {
		c, ok := store.byID[id]
		if !ok {
			t.Fatalf("fixture missing challenge %d", id)
		}
		if id > 1 && c.PrevCode != prev {
			t.Errorf("challenge %d prevCode %q != challenge %d flag %q", id, c.PrevCode, id-1, prev)
		}
		prev = c.Flag
	}
}
