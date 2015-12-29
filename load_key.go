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

func LoadKey(keyfile string) (Key, error) {
	var key internal.PrivateKey
	f, err := os.Open(keyfile)
	if err != nil {
		if !os.IsNotExist(err) {
			return Key{key}, err
		}
		key, err := NewKey()
		b, err := json.Marshal(key.k)
		if err != nil {
			return key, err
		}
		f, err = os.Create(keyfile)
		if err != nil {
			return key, err
		}
		_, err = f.Write(b)
		return key, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return Key{key}, err
	}
	err = json.Unmarshal(b, &key)
	return Key{key}, err
}
