package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/AutoRoute/node/internal"
)

type Key struct {
	k internal.PrivateKey
}

func (k Key) String() string {
	return fmt.Sprintf("Key{%v}", k.k.PublicKey().Hash())
}

func NewKey() (Key, error) {
	key, err := internal.NewECDSAKey()
	return Key{key}, err
}

// Creates a new keyfile
// Args:
//	keyfile: name of the keyfile
// Returns:
//	file, pirvate key, error
func CreateKey(keyfile string) (*os.File, internal.PrivateKey, error) {
	key, err := NewKey()
	if err != nil {
		return nil, key.k, err
	}
	b, err := json.Marshal(key.k)
	if err != nil {
		return nil, key.k, err
	}
	f, err := os.Create(keyfile)
	if err != nil {
		return f, key.k, err
	}
	_, err = f.Write(b)
	if err != nil {
		return f, key.k, err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return f, key.k, err
	}
	return f, key.k, nil
}

// Loads a key from a keyfile
// Args:
//	keyfile: name of the keyfile
// Returns:
//	private key (type Key)
func LoadKey(keyfile string) (Key, error) {
	var key internal.PrivateKey
	f, err := os.Open(keyfile)
	if err != nil {
		if !os.IsNotExist(err) {
			return Key{key}, err
		}
		f, key, err = CreateKey(keyfile)
		if err != nil {
			return Key{key}, err
		}
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return Key{key}, err
	}
	err = json.Unmarshal(b, &key)
	return Key{key}, err
}
