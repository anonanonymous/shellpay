package user

import (
	"crypto/sha256"
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
	Username   string
	Verifier   string
	IH         string // srp identity
	Email      string
	TOTPKey    string // 2FA secret key
	PrivateKey []byte // HMAC private key
}

// NewUser - creates a new user
func NewUser(uname, pwd, email string) (*User, error) {
	if len(uname) < 1 || len(pwd) < 1 {
		return nil, errors.New("Username/Password too short")
	}

	if uname[0] == ' ' || uname[len(uname)-1] == ' ' {
		return nil, errors.New("Leading/Trailing spaces in username")
	}

	if len(email) != 0 {
		matched, err := regexp.MatchString("(\\w+?@\\w+?\\x2E.+)", email)
		if err != nil || !matched {
			return nil, errors.New("Invalid Email")
		}
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

	private := sha256.Sum256([]byte(vfr))

	return &User{
		Username:   uname,
		Verifier:   vfr,
		IH:         ih,
		Email:      email,
		PrivateKey: private[:],
	}, nil
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
func (u *User) DisableTwoFactor() error {
	if !u.TwoFactorEnabled() {
		return errors.New("2FA already disabled")
	}

	u.TOTPKey = ""
	return nil
}

// EnableTwoFactor - creates a TOTP secret for the user
func (u *User) EnableTwoFactor() error {
	if u.TwoFactorEnabled() {
		return errors.New("2FA already enabled")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      domainName,
		AccountName: ":",
		SecretSize:  otpSecretSize,
	})

	if err != nil {
		return err
	}

	u.TOTPKey = key.Secret()
	return nil
}
