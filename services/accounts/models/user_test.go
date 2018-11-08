package user

import (
	"testing"
)

type userDB struct {
	users map[string]User
}

func (db userDB) insert(usr User) {
	db.users[usr.IH] = usr
}

func (db userDB) query(ident string) (User, bool) {
	u, ok := db.users[ident]
	return u, ok
}

func TestUser(t *testing.T) {
	var err error
	var u *User

	goodUsers := [][]string{
		{"username", "password", "email@email.com"},
		{"username", "password", ""},
		{"user name", "pass word", ""},
	}
	badUsers := [][]string{
		{"", "password", "gmail@gmail.com"},
		{"username", "", "cl@o.k"},
		{"username", "pwd", "benisXDDD"},
		{"", "", ""},
		{"  XD", "ok", ""},
		{"XD ", "ok", ""},
	}

	for _, creds := range goodUsers {
		u, err = NewUser(creds[0], creds[1], creds[2])
		if err != nil {
			t.Errorf("Expected: <nil> got: <%s>", err)
		} else if u.Username != creds[0] {
			t.Errorf("Expected: %s got: %s", creds[0], u.Username)
		} else if len(u.IH) != 64 {
			t.Error("Bad User Identity value")
		} else if len(u.APIToken) != 64 {
			t.Error("Bad API Token")
		}

		ok, err := u.Verify(creds[1])
		if !ok || err != nil {
			t.Errorf("User can't login")
		}

		ok, err = u.Verify("not the password")
		if ok || err == nil {
			t.Errorf("Hackerman can login")
		}
	}

	for _, creds := range badUsers {
		_, err := NewUser(creds[0], creds[1], creds[2])
		if err == nil {
			t.Error("Bad user created")
		}
	}

	u, _ = NewUser("james", "bond", "")
	if err := u.EnableTwoFactor(); err != nil || len(u.TOTPKey) == 0 {
		t.Fail()
	}
	if err := u.EnableTwoFactor(); err == nil || len(u.TOTPKey) == 0 {
		t.Fail()
	}
	if err := u.DisableTwoFactor(); err != nil || len(u.TOTPKey) != 0 {
		t.Fail()
	}
	if err := u.DisableTwoFactor(); err == nil || len(u.TOTPKey) != 0 {
		t.Fail()
	}
}
