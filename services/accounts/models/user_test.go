package user

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

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
		} else if len(u.TOTPKey) != 0 {
			t.Error("TOTP Secret not empty")
		}

		ok, err := u.Verify(creds[1])
		if !ok || err != nil {
			t.Errorf("User can't login")
		}

		ok, err = u.Verify("not the password")
		if ok || err == nil {
			t.Errorf("Hackerman can login")
		}

		jsMap := u.Jsonify()
		if jsMap["username"] != u.Username {
			t.Error("Username does not match")
		}
		if jsMap["identity"] != u.IH {
			t.Error("Identities do not match")
		}
		if jsMap["totpKey"] != u.TOTPKey {
			t.Error("TOTP Key does not match")
		}
		if jsMap["verifier"] != u.Verifier {
			t.Error("Verifier does not match")
		}
		if jsMap["email"] != u.Email {
			t.Error("Email does not match")
		}
	}

	for _, creds := range badUsers {
		_, err := NewUser(creds[0], creds[1], creds[2])
		if err == nil {
			t.Error("Bad user created")
		}
	}

	u, _ = NewUser("james", "bond", "")
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      domainName,
		AccountName: ":",
		SecretSize:  otpSecretSize,
	})

	secret := key.Secret()
	code, _ := totp.GenerateCode(secret, time.Now())
	if err := u.EnableTwoFactor(secret, code); err != nil || len(u.TOTPKey) != 64 {
		t.Fail()
	}

	code, _ = totp.GenerateCode(secret, time.Now())
	if err := u.EnableTwoFactor(secret, code); err == nil || len(u.TOTPKey) == 0 {
		t.Fail()
	}

	code, _ = totp.GenerateCode(secret, time.Now())
	if err := u.DisableTwoFactor(code); err != nil || len(u.TOTPKey) != 0 {
		t.Fail()
	}

	code, _ = totp.GenerateCode(secret, time.Now())
	if err := u.DisableTwoFactor(code); err == nil || len(u.TOTPKey) != 0 {
		t.Fail()
	}
}
