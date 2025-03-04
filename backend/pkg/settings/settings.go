package settings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

// Settings represents the main configuration structure
type Settings struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"Server"`
	Pocketbase struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"Pocketbase"`
	Mailer struct {
		Service     string `yaml:"service"`
		Host        string `yaml:"host"`
		Port        int    `yaml:"port"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		FromAddress string `yaml:"from_address"`
		FromName    string `yaml:"from_name"`
	} `yaml:"Mailer"`
	AppUrl         string `yaml:"AppUrl"`
	IsInitialized  bool   `yaml:"IsInitialized"`
	JWTSecret      string `yaml:"JWTSecret"`
	MaxLabbookSize int64  `yaml:"MaxLabbookSize"` // In MB
}

// LoadSettings reads and parses the settings.yml file
func LoadSettings(filepath string) (*Settings, error) {
	settings := &Settings{}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// Save changes to settings file
func (s *Settings) Save(filepath string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

func UpdatePBSettings(settings Settings) (err error) {
	url := "http://127.0.0.1:8090/api/settings/smtp"

	data := map[string]interface{}{
		"host":         settings.Mailer.Host,
		"port":         settings.Mailer.Port,
		"username":     settings.Mailer.Username,
		"password":     settings.Mailer.Password,
		"from_address": settings.Mailer.FromAddress,
		"from_name":    settings.Mailer.FromName,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("faile	d to create request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update settings: status code %d", resp.StatusCode)
	}
	log.Println("PocketBase settings updated successfully")
	return nil
}
