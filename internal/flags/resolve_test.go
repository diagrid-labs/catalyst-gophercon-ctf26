package flags

import "testing"

func TestResolve(t *testing.T) {
	store, err := New([]Challenge{
		{ID: 1, Tool: "board_tour", PrevCode: "code1", Flag: "gc26{one}"},
		{ID: 2, Tool: "restore_power", PrevCode: "gc26{one}", Flag: "gc26{two}"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tests := []struct {
		name     string
		id       int
		code     string
		wantFlag string
		wantOK   bool
	}{
		{"correct code", 1, "code1", "gc26{one}", true},
		{"wrong code", 1, "nope", "", false},
		{"empty code", 1, "", "", false},
		{"chained correct", 2, "gc26{one}", "gc26{two}", true},
		{"chained wrong", 2, "code1", "", false},
		{"unknown challenge", 9, "code1", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, ok := store.Resolve(tt.id, tt.code)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if flag != tt.wantFlag {
				t.Errorf("flag = %q, want %q", flag, tt.wantFlag)
			}
		})
	}
}
