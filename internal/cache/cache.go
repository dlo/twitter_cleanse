package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	baseDir string
}

func New(baseDir string) *Cache {
	return &Cache{
		baseDir: baseDir,
	}
}

func (c *Cache) generateKey(endpoint string, params map[string]string) string {
	key := endpoint
	for k, v := range params {
		key += fmt.Sprintf("%s=%s", k, v)
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

func (c *Cache) Get(endpoint string, params map[string]string, result interface{}) error {
	if err := os.MkdirAll(c.baseDir, 0755); err != nil {
		return err
	}

	key := c.generateKey(endpoint, params)
	filePath := filepath.Join(c.baseDir, key+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}

func (c *Cache) Set(endpoint string, params map[string]string, data interface{}) error {
	if err := os.MkdirAll(c.baseDir, 0755); err != nil {
		return err
	}

	key := c.generateKey(endpoint, params)
	filePath := filepath.Join(c.baseDir, key+".json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
