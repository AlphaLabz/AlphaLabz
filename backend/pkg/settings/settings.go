package settings

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Pocketbase struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Mailer struct {
	Service     string `yaml:"service"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	FromAddress string `yaml:"from_address"`
	FromName    string `yaml:"from_name"`
}

// Settings represents the main configuration structure
type Settings struct {
	Server        Server     `yaml:"Server"`
	Pocketbase    Pocketbase `yaml:"Pocketbase"`
	Mailer        Mailer     `yaml:"Mailer"`
	AppUrl        string     `yaml:"AppUrl"`
	IsInitialized bool       `yaml:"IsInitialized"`
	JWTSecret     string     `yaml:"JWTSecret"`
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
