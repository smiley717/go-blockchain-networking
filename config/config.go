package config

import (
	_ "embed"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

// peerSeed is a single peer entry from the config's seed list
type peerSeed struct {
	PublicKey string   `yaml:"PublicKey"` // Public key = peer ID. Hex encoded.
	Address   []string `yaml:"Address"`   // IP:Port
}

type Config struct {
	PrivateKey string     `yaml:"PrivateKey"` // The Private Key, hex encoded so it can be copied manually
	SeedList   []peerSeed `yaml:"SeedList"`   // Initial peer seed list
}

//go:embed "config.yaml"
var ConfigDefault []byte

// LoadConfig reads the YAML configuration file and unmarshall it into the provided structure.
// If the config file does not exist or is empty, it will fall back to the default config which is hardcoded.
func LoadConfig(Filename string, ConfigOut interface{}) (status int, err error) {
	var configData []byte

	// check if the file is non existent or empty
	stats, err := os.Stat(Filename)
	if err != nil && os.IsNotExist(err) || err == nil && stats.Size() == 0 {
		configData = ConfigDefault
	} else if err != nil {
		return ExitErrorConfigAccess, err
	} else if configData, err = ioutil.ReadFile(Filename); err != nil {
		return ExitErrorConfigRead, err
	}

	// parse the config
	err = yaml.Unmarshal(configData, ConfigOut)
	if err != nil {
		return ExitErrorConfigParse, err
	}

	return ExitSuccess, nil
}
