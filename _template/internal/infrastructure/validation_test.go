package infrastructure

import "testing"

func TestIsValidEmail_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"user.name+tag@example.co.uk",
		"user_name@example.org",
		"user-name@example.io",
		"a@b.co",
	}
	for _, c := range cases {
		if !IsValidEmail(c) {
			t.Errorf("expected valid: %q", c)
		}
	}
}

func TestIsValidEmail_Invalid(t *testing.T) {
	cases := []string{
		"",
		"not-an-email",
		"@example.com",
		"user@",
		"user@.com",
		"user@example.",
		"user@example",
		"user @example.com",
	}
	for _, c := range cases {
		if IsValidEmail(c) {
			t.Errorf("expected invalid: %q", c)
		}
	}
}
