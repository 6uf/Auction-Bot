package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func (s *Channels) LoadState() {
	data, err := ReadFile("config.json")
	if err != nil {
		fmt.Print("No config file found, loading one.\n\n")

		fmt.Print("Disord Bot Key: ")
		fmt.Scan(&Data.Key)

		s.LoadFromFile()
		s.SaveConfig()
		os.Exit(0)
	}

	json.Unmarshal([]byte(data), s)
	s.LoadFromFile()
}

func (c *Channels) LoadFromFile() {
	// Load a config file

	jsonFile, err := os.Open("config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		jsonFile, _ = os.Create("config.json")
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &c)
}

func (config *Channels) SaveConfig() {
	WriteFile("config.json", string(config.ToJson()))
}

func (s *Channels) ToJson() []byte {
	b, _ := json.MarshalIndent(s, "", "  ")
	return b
}

func WriteFile(path string, content string) {
	ioutil.WriteFile(path, []byte(content), 0644)
}

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
