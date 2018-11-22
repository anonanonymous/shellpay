package user

import (
	"crypto/subtle"
	"errors"
	"regexp"

	"github.com/opencoff/go-srp"
	"github.com/pquerna/otp/totp"
)

const (
	nBits         = 1024 // prime-field size for srp verifier
	otpSecretSize = 40
	domainName    = "shellpay.idk"
)

// User - A User object
type User struct {
	ID       string
	Username string
	Verifier string
	IH       string // srp identity
	Email    string
	TOTPKey  string // 2FA secret key
}

// NewUser - creates a new user
func NewUser(uname, pwd, email string) (*User, error) {
	usr := &User{}
	if len(uname) < 1 || len(pwd) < 1 {
		return nil, errors.New("Username/Password too short")
	}

	if uname[0] == ' ' || uname[len(uname)-1] == ' ' {
		return nil, errors.New("Leading/Trailing spaces in username")
	}

	if !usr.SetEmail(email) {
		return nil, errors.New("Invalid Email")
	}

	srpEnv, err := srp.New(nBits)
	if err != nil {
		return nil, err
	}

	vh, err := srpEnv.Verifier([]byte(uname), []byte(pwd))
	if err != nil {
		return nil, err
	}

	ih, vfr := vh.Encode()

	usr.Username = uname
	usr.Verifier = vfr
	usr.IH = ih

	return usr, nil
}

// Verify - checks if password is valid
func (u User) Verify(pwd string) (bool, error) {
	srpEnv, err := srp.New(nBits)
	if err != nil {
		return false, err
	}

	client, err := srpEnv.NewClient([]byte(u.Username), []byte(pwd))
	if err != nil {
		return false, err
	}

	creds := client.Credentials()
	ih, A, err := srp.ServerBegin(creds)
	if err != nil {
		return false, err
	}

	if ih != u.IH {
		return false, errors.New("Identities are not equal")
	}

	s, verif, err := srp.MakeSRPVerifier(u.Verifier)
	if err != nil {
		return false, err
	}

	svr, err := s.NewServer(verif, A)
	if err != nil {
		return false, err
	}

	creds = svr.Credentials()
	cauth, err := client.Generate(creds)
	if err != nil {
		return false, err
	}

	proof, ok := svr.ClientOk(cauth)
	if err != nil {
		return false, err
	}

	if !ok || !client.ServerOk(proof) {
		return false, errors.New("Invalid Verifier")
	}

	if 1 != subtle.ConstantTimeCompare(client.RawKey(), svr.RawKey()) {
		return false, errors.New("Invalid Keys")
	}

	return true, nil
}

// TwoFactorEnabled  - determines if 2FA is enabled
func (u User) TwoFactorEnabled() bool {
	return len(u.TOTPKey) != 0
}

// DisableTwoFactor - removes TOTP secret from the user
func (u *User) DisableTwoFactor(code string) error {
	if !u.TwoFactorEnabled() {
		return errors.New("2FA already disabled")
	}

	if totp.Validate(code, u.TOTPKey) {
		u.TOTPKey = ""
		return nil
	}
	return errors.New("Invalid code")
}

// EnableTwoFactor - creates a TOTP secret for the user
func (u *User) EnableTwoFactor(secret, code string) error {
	if u.TwoFactorEnabled() {
		return errors.New("2FA already enabled")
	}
	if totp.Validate(code, secret) {
		u.TOTPKey = secret
		return nil
	}
	return errors.New("Invalid code")
}

// SetEmail - sets the users email
func (u *User) SetEmail(email string) bool {
	if len(email) != 0 {
		matched, err := regexp.MatchString("(\\w+?@\\w+?\\x2E.+)", email)
		if err != nil || !matched {
			return false
		}
	}
	u.Email = email
	return true
}

// Jsonify - returns a json representation of the user
func (u User) Jsonify() map[string]string {
	return map[string]string{
		"id":       u.ID,
		"username": u.Username,
		"verifier": u.Verifier,
		"email":    u.Email,
		"identity": u.IH,
		"totpKey":  u.TOTPKey,
	}
}
