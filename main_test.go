package main

import "testing"

// unit test for CheckDate function
func TestCheckDate(t *testing.T) {
	// table of test cases
	tests := []struct {
		input string
		want  bool
	}{
		{"2020-01-01", true},
		{"20-2001-01", false},
		{"2020-01-23", true},
		{"2020-20-01", false},
	}

	// run test cases
	for _, test := range tests {
		if got := CheckDate(test.input); got != test.want {
			t.Errorf("CheckDate(%q) = %v, want %v", test.input, got, test.want)
		}
	}

}
