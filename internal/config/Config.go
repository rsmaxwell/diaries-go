package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MqttConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Go struct {
	Driver string `json:"driver"`
}

type Jdbc struct {
	Dbms string `json:"dbms"`
}

type DBConfig struct {
	Go       Go     `json:"go"`
	Jdbc     Jdbc   `json:"jdbc"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Mqtt MqttConfig `json:"mqtt"`
	Db   DBConfig   `json:"db"`
}

func (c *MqttConfig) GetServer() string {
	return fmt.Sprintf("mqtt://%s:%d", c.Host, c.Port)
}

func (c *DBConfig) DriverName() string {
	return c.Go.Driver
}

func (c *DBConfig) ConnectionString(database string) string {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s",
		c.Host, c.Port, c.Username, c.Password)

	if database != "" {
		return fmt.Sprintf("%s dbname=%s", connectionString, database)
	}

	return connectionString
}

// func (c *DBConfig) GetJdbcUrl() string {
// 	return fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
// 		c.Host, c.Port, c.Username, c.Password)
// }

// func (c *DBConfig) GetJdbcUrlWithDatabase() string {
// 	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
// 		c.Host, c.Port, c.Username, c.Password, c.Database)
// }

func Read() (*Config, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(dirname, ".diaries", "responder.json")

	configFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Print() {

	configBytes, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(configBytes))
}
