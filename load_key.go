package node

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func LoadKey(keyfile string) (PrivateKey, error) {
	var key PrivateKey
	f, err := os.Open(keyfile)
	if err != nil {
		if !os.IsNotExist(err) {
			return key, err
		}
		key, err = NewECDSAKey()
		if err != nil {
			return key, err
		}
		b, err := json.Marshal(key)
		if err != nil {
			return key, err
		}
		f, err = os.Create(keyfile)
		if err != nil {
			return key, err
		}
		_, err = f.Write(b)
		if err != nil {
			return key, err
		}
		_, err = f.Seek(0, 0)
		if err != nil {
			return key, err
		}
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return key, err
	}
	err = json.Unmarshal(b, &key)
	if err != nil {
		return key, err
	}
	return key, err
}
