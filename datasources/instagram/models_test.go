package instagram

import "testing"

func TestProfileStringMapGet(t *testing.T) {
	m := profileStringMap{
		"Username":       instaProfileData{Value: "_sam_mie_"},
		"E-Mail-Adresse": instaProfileData{Value: "samuel.moede@outlook.de"},
		"Geschlecht":     instaProfileData{Value: "male"},
	}

	tests := []struct {
		name         string
		canonicalKey string
		wantValue    string
	}{
		{
			name:         "exact canonical (English) key present",
			canonicalKey: "Username",
			wantValue:    "_sam_mie_",
		},
		{
			name:         "German alias for a canonical key not present verbatim",
			canonicalKey: "Email",
			wantValue:    "samuel.moede@outlook.de",
		},
		{
			name:         "German alias for Gender",
			canonicalKey: "Gender",
			wantValue:    "male",
		},
		{
			name:         "no match in either canonical or alias form",
			canonicalKey: "Website",
			wantValue:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.get(tt.canonicalKey).Value
			if got != tt.wantValue {
				t.Fatalf("get(%q) = %q, want %q", tt.canonicalKey, got, tt.wantValue)
			}
		})
	}
}

func TestProfileStringMapGetPrefersCanonicalOverAlias(t *testing.T) {
	// If a future export somehow has both keys, the canonical (English) one
	// should win, since it's checked first.
	m := profileStringMap{
		"Username":     instaProfileData{Value: "english-value"},
		"Benutzername": instaProfileData{Value: "german-value"},
	}
	if got := m.get("Username").Value; got != "english-value" {
		t.Fatalf("get(%q) = %q, want %q", "Username", got, "english-value")
	}
}
