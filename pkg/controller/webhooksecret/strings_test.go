package webhooksecret

import (
	"reflect"
	"testing"
)

func TestContainsString(t *testing.T) {
	list := []string{"one", "two"}
	containsTests := []struct {
		key  string
		want bool
	}{
		{"one", true},
		{"unknown", false},
	}

	for _, tt := range containsTests {
		if v := containsString(list, tt.key); v != tt.want {
			t.Errorf("looking up key %#v got %v, want %v", tt.key, v, tt.want)
		}
	}
}

func TestRemoveString(t *testing.T) {
	list := []string{"one", "two"}

	got := removeString(list, "two")
	want := []string{"one"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#vm want %#v", got, want)
	}
}
