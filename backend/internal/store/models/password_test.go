package models

import "testing"

func TestSetAndCheckPassword(t *testing.T) {
	user := &User{}
	plain := "password123"

	if err := user.SetPassword(plain); err != nil {
		t.Fatalf("expected no error setting password, got %v", err)
	}

	if len(user.Password) == 0 {
		t.Fatal("expected password to not be empty")
	}

	if !user.CheckPassword(plain) {
		t.Fatalf("expected CheckPassword to return true for correct password")
	}
}

func TestCheckPasswordWithWrongPassword(t *testing.T) {
	user := &User{}

	if err := user.SetPassword("correctPassword"); err != nil {
		t.Fatalf("expected no error setting password, got %v", err)
	}

	if user.CheckPassword("wrongPassword") {
		t.Fatal("expected Checkpassword to return false for wrong password")
	}
}

func TestPasswordIsHashed(t *testing.T) {
	user := &User{}
	plain := "password123"

	if err := user.SetPassword(plain); err != nil {
		t.Fatalf("expected no error setting password, got %v", err)
	}

	if string(user.Password) == plain {
		t.Fatal("expected password to be hashed, not stored as plain text")
	}
}