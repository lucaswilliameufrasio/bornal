package pass

import "testing"

func TestParsePinentryReturnsOnlyTheSecretLine(t *testing.T) {
	value, err := ParsePinentry("OK\nD correct-horse\nOK\n")
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != "correct-horse" {
		t.Fatalf("valor %q", value)
	}
	if _, err := ParsePinentry("ERR 83886179 cancelled\n"); err != ErrCancel {
		t.Fatalf("cancel %v", err)
	}
}

func TestPairRejectsAMismatch(t *testing.T) {
	n := 0
	_, err := Pair(func(string) ([]byte, error) {
		n++
		if n == 1 {
			return []byte("one"), nil
		}
		return []byte("two"), nil
	}, "a", "b")
	if err != ErrMismatch {
		t.Fatalf("err %v", err)
	}
}
