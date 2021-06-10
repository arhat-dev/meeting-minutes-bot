package conf

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"arhat.dev/meeting-minutes-bot/pkg/storage"
)

type StorageConfig struct {
	Driver        string `json:"driver" yaml:"driver"`
	MIMEMatch     string `json:"mimeMatch" yaml:"mimeMatch"`
	MaxUploadSize uint64 `json:"maxUploadSize" yaml:"maxUploadSize"`

	Config interface{} `json:"config" yaml:"config"`
}

func (c *StorageConfig) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})

	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	c.Driver, c.MIMEMatch, c.MaxUploadSize, c.Config, err = unmarshalStorageConfig(m)
	if err != nil {
		return err
	}

	return nil
}

func (c *StorageConfig) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]interface{})

	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, m)
	if err != nil {
		return err
	}

	c.Driver, c.MIMEMatch, c.MaxUploadSize, c.Config, err = unmarshalStorageConfig(m)
	if err != nil {
		return err
	}

	return nil
}

func unmarshalStorageConfig(
	m map[string]interface{},
) (driver, mimeMatch string, maxSize uint64, config interface{}, err error) {
	n, ok := m["driver"]
	if !ok {
		return
	}

	driver, ok = n.(string)
	if !ok {
		err = fmt.Errorf("storage driver name must be a string")
		return
	}

	match, ok := m["mimeMatch"]
	if ok {
		mimeMatch, ok = match.(string)
		if !ok {
			err = fmt.Errorf("mimeMatch must be a string")
			return
		}
	}

	switch s := m["maxUploadSize"].(type) {
	case float64:
		maxSize = uint64(s)
	case float32:
		maxSize = uint64(s)
	case int64:
		maxSize = uint64(s)
	case int32:
		maxSize = uint64(s)
	case int16:
		maxSize = uint64(s)
	case int8:
		maxSize = uint64(s)
	case int:
		maxSize = uint64(s)
	case uint64:
		maxSize = s
	case uint32:
		maxSize = uint64(s)
	case uint16:
		maxSize = uint64(s)
	case uint8:
		maxSize = uint64(s)
	case uint:
		maxSize = uint64(s)
	case nil:
	default:
		err = fmt.Errorf("invaid type of maxUploadSize, expecting an integer: %T", s)
		return
	}

	config, err = storage.NewConfig(driver)
	if err != nil {
		return driver, "", 0, nil, nil
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
			err = fmt.Errorf("failed to get storage config bytes: %w", err)
			return
		}
	}

	dec := yaml.NewDecoder(bytes.NewReader(configData))
	dec.KnownFields(true)
	err = dec.Decode(config)

	return
}
