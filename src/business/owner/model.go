package owner

import (
	"encoding/json"
	"github.com/Power-LAB/PeerVault/crypto"
	"github.com/Power-LAB/PeerVault/database"
	"github.com/Power-LAB/PeerVault/identity"
	"go.etcd.io/bbolt"
	"net/http"
)

const (
	PasswordPolicyNone             = iota // Password never required
	PasswordPolicyAlwaysRequired   = iota // Password always ask
	PasswordPolicyOnlyWhenExposure = iota // Password required when expose secret
)

type Owner struct {
	// Public
	QmPeerId string
	Nickname string
	DeviceName string
	UnlockCode string   //`json:"-"`    WE CANT USE IT, BECAUSE WE ALSO CONVERT IN JSON ON BBOLT
	AskPassword int     // PasswordPolicyNone | PasswordPolicyAlwaysRequired | PasswordPolicyOnlyWhenExposure
}

func IsOwnerExist() (bool, error) {
	db, err := database.GetConnection()
	if err != nil {
		return false, err
	}

	exist := false
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("owner"))
		if b == nil {
			return nil
		}
		buf := b.Get([]byte("buf"))
		if len(buf) > 0 {
			exist = true
		}
		return nil
	})

	return exist, err
}

func (o *Owner) GetIdentity() (identity.PeerIdentity, error) {
	return identity.GetIdentity(o.QmPeerId)
}

func (o *Owner) FetchOwner() error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("owner"))
		buf := b.Get([]byte("buf"))
		err := json.Unmarshal(buf, o)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *Owner) PutOwner() error {
	db, err := database.GetConnection()
	if err != nil {
		return err
	}

	keychain := crypto.Keychain{}
	err = keychain.CreateOrOpen()
	if err != nil {
		return err
	}
	log.Debug("update UnlockCode", o.UnlockCode)
	err = keychain.Put("UnlockCode", []byte(o.UnlockCode), "OwnerCode", true)
	// erase code because we only want it into keychain, not bbolt
	o.UnlockCode = ""

	// Marshal Owner into bytes.
	buf, err := json.Marshal(&o)
	if err != nil {
		return err
	}

	code, _ := keychain.Get("UnlockCode", "OwnerCode")
	log.Debugf("assert code %s", code)

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("owner"))
		if b == nil {
			log.Debug("Bucket owner is nil")
			b2, err := tx.CreateBucket([]byte("owner"))
			if err != nil {
				log.Debug("Bucket Creation error")
				return err
			}
			b = b2
		}
		return b.Put([]byte("buf"), buf)
	})
}

// Verification of the password of the owner
// We also check that the owner exist, if not, we return FALSE as we assume is not authorized by definition
func PasswordVerification(r *http.Request, exposure bool) bool {
	exist, err := IsOwnerExist()
	if err != nil || exist == false {
		return false
	}
	o := &Owner{}
	err = o.FetchOwner()
	if err != nil {
		return false
	}

	// Always authorized when password is disabled
	if o.AskPassword == PasswordPolicyNone {
		log.Debug("PasswordVerification PasswordPolicyNone")
		return true
	}

	// Authorized when we are not exposing secret and password policy is exposure only
	if !exposure && o.AskPassword == PasswordPolicyOnlyWhenExposure {
		log.Debug("PasswordVerification PasswordPolicyOnlyWhenExposure")
		return true
	}

	keychain := crypto.Keychain{}
	err = keychain.CreateOrOpen()
	if err != nil {
		log.Error(err)
		return false
	}
	code, err := keychain.Get("UnlockCode", "OwnerCode")
	if err != nil {
		log.Error(err)
		return false
	}
	log.Debugf("assert code %s = %s", code, r.Header.Get("X-OWNER-CODE"))

	return string(code) == r.Header.Get("X-OWNER-CODE")
}