package config

import (
	"bytes"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

var defaultPath string

func init() {
	home, _ := homedir.Dir()
	defaultPath = path.Join(home, ".helium-dashboard", "config.json")
}

type fileData struct {
	Addresses []string `json:"addresses"`
}

type Data struct {
	lock sync.RWMutex

	addressIndex map[string]struct{}
	filePath     string
}

func FromFile(path *string) (*Data, error) {
	if path == nil {
		path = &defaultPath
	}

	file, err := ioutil.ReadFile(*path)
	if err != nil {
		return nil, err
	}

	var fileData *fileData
	err = json.Unmarshal(file, &fileData)
	if err != nil {
		return nil, err
	}

	data := new(Data)
	data.addressIndex = make(map[string]struct{})
	for _, address := range fileData.Addresses {
		data.addressIndex[address] = struct{}{}
	}
	data.filePath = *path

	return data, nil
}

func (d *Data) Save() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	newAddressList := make([]string, 0, len(d.addressIndex))
	for address := range d.addressIndex {
		newAddressList = append(newAddressList, address)
	}

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(d.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0)
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	_, err = io.Copy(file, bytes.NewReader(b))
	return err
}

func (d *Data) AddAddress(address string) {
	d.lock.Lock()
	d.addressIndex[address] = struct{}{}
	d.lock.Unlock()
}

func (d *Data) DeleteAddress(address string) {
	d.lock.Lock()
	delete(d.addressIndex, address)
	d.lock.Unlock()
}

func (d *Data) Addresses() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	var addresses []string
	for address := range d.addressIndex {
		addresses = append(addresses, address)
	}

	return addresses
}