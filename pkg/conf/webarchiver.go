package conf

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

type WebArchiverConfig struct {
	Driver string      `json:"driver" yaml:"driver"`
	Config interface{} `json:"config" yaml:"config"`
}

func (c *WebArchiverConfig) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	c.Driver, c.Config, err = unmarshalWebArchiverConfig(m)
	if err != nil {
		return err
	}

	return nil
}

func (c *WebArchiverConfig) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]interface{})

	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, m)
	if err != nil {
		return err
	}

	c.Driver, c.Config, err = unmarshalWebArchiverConfig(m)
	if err != nil {
		return err
	}

	return nil
}

func unmarshalWebArchiverConfig(m map[string]interface{}) (driver string, config interface{}, err error) {
	n, ok := m["driver"]
	if !ok {
		return
	}

	driver, ok = n.(string)
	if !ok {
		err = fmt.Errorf("web archiver name must be a string")
		return
	}

	config, err = webarchiver.NewConfig(driver)
	if err != nil {
		return driver, nil, nil
	}

	configRaw, ok := m["config"]
	if !ok {
		return
	}

	var configData []byte
	switch d := configRaw.(type) {
	case []byte:
		configData = d
	case string:
		configData = []byte(d)
	default:
		configData, err = yaml.Marshal(d)
		if err != nil {
			err = fmt.Errorf("failed to get web archiver config bytes: %w", err)
			return
		}
	}

	dec := yaml.NewDecoder(bytes.NewReader(configData))
	dec.KnownFields(true)
	err = dec.Decode(config)
	if err != nil {
		return
	}

	return driver, config, nil
}
