package owner

import (
	"encoding/json"
	"fmt"
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
	// Private
	identity identity.PeerIdentityJson
	// Public
	Nickname string
	DeviceName string
	Code string 		//`json:"-"`    WE CANT USE IT, BECAUSE WE ALSO CONVERT IN JSON ON BBOLT
	AskPassword int // PasswordPolicyNone | PasswordPolicyAlwaysRequired | PasswordPolicyOnlyWhenExposure
}

func IsOwnerExist() (bool, error) {
	db, err := database.Open()
	if err != nil {
		return false, err
	}
	defer db.Close()

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

func (o *Owner) GetIdentity() identity.PeerIdentityJson {
	return o.identity
}

func (o *Owner) FetchOwner() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("owner"))
		buf := b.Get([]byte("buf"))
		err := json.Unmarshal(buf, o)
		if err != nil {
			return err
		}

		ownerIdentity := identity.PeerIdentityJson{}
		ownerIdentityJson := b.Get([]byte("identity"))
		err2 := json.Unmarshal(ownerIdentityJson, &ownerIdentity)
		if err2 != nil {
			return err2
		}
		o.identity = ownerIdentity

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *Owner) PutOwner() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Marshal Owner into bytes.
	buf, err := json.Marshal(&o)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("owner"))
		if b == nil {
			fmt.Println("Bucket owner is nil")
			b2, err := tx.CreateBucket([]byte("owner"))
			if err != nil {
				fmt.Println("Bucket Creation error")
				return err
			}
			b = b2
		}
		return b.Put([]byte("buf"), buf)
	})
}

func (o *Owner) SaveIdentity() error {
	db, err := database.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Marshal Owner Identity into bytes.
	buf, err := json.Marshal(&o.identity)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		var b *bbolt.Bucket
		b = tx.Bucket([]byte("owner"))
		if b == nil {
			fmt.Println("Bucket owner is nil")
			b2, err := tx.CreateBucket([]byte("owner"))
			if err != nil {
				fmt.Println("Bucket Creation error")
				return err
			}
			b = b2
		}
		return b.Put([]byte("identity"), buf)
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
		fmt.Println("PasswordVerification PasswordPolicyNone")
		return true
	}

	// Authorized when we are not expose secret and password policy is exposure only
	if !exposure && o.AskPassword == PasswordPolicyOnlyWhenExposure {
		fmt.Println("PasswordVerification PasswordPolicyOnlyWhenExposure")
		return true
	}

	return o.Code == r.Header.Get("X-OWNER-CODE")
}