package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Info struct {
	ChannelID string       `json:"channelID"`
	History   []History    `json:"bidhistory"`
	StartBid  int64        `json:"startbid"`
	Name      string       `json:"name"`
	Info      string       `json:"info"`
	MessageID string       `json:"messageid"`
	Claimed   bool         `json:"claimed"`
	Roles     RoleSpecific `json:"roleinfo"`
}

type Guilds struct {
	Token   string `json:"discordkey"`
	GuildID []struct {
		GuildID string
		IDs     []string `json:"verified"`
		Bans    []string `json:"banned"`
		Data    []Info
	} `json:"guildid"`
}

func (s *Guilds) LoadStateClient() {
	data, err := ReadFile("database.json")
	if err != nil {
		s.LoadFromFileClient()
		s.SaveConfigClient()
		os.Exit(0)
	}

	json.Unmarshal([]byte(data), s)
	s.LoadFromFileClient()
}

func (c *Guilds) LoadFromFileClient() {
	// Load a config file

	jsonFile, err := os.Open("database.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		jsonFile, _ = os.Create("database.json")
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &c)
}

func (config *Guilds) SaveConfigClient() {
	WriteFile("database.json", string(config.ToJsonClient()))
}

func (s *Guilds) ToJsonClient() []byte {
	b, _ := json.MarshalIndent(s, "", "  ")
	return b
}

func (s *Guilds) GetGuildData(ID string) struct {
	GuildID string
	IDs     []string "json:\"verified\""
	Bans    []string "json:\"banned\""
	Data    []Info
} {
	if ok, info := s.CheckGuild(ID); ok {
		return info
	}
	return struct {
		GuildID string
		IDs     []string "json:\"verified\""
		Bans    []string "json:\"banned\""
		Data    []Info
	}{}
}

func (s *Guilds) UpdateInput(data struct {
	GuildID string
	IDs     []string "json:\"verified\""
	Bans    []string "json:\"banned\""
	Data    []Info
}) {
	for i, client := range s.GuildID {
		if data.GuildID == client.GuildID {
			s.GuildID[i] = data
		}
	}
}

func (s *Guilds) CheckGuild(ID string) (bool, struct {
	GuildID string
	IDs     []string "json:\"verified\""
	Bans    []string "json:\"banned\""
	Data    []Info
}) {
	for _, data := range s.GuildID {
		if data.GuildID == ID {
			return true, data
		}
	}
	return false, struct {
		GuildID string
		IDs     []string "json:\"verified\""
		Bans    []string "json:\"banned\""
		Data    []Info
	}{}
}

func (s *Guilds) AddGuild(ID string) {
	if ok, info := s.CheckGuild(ID); !ok {
		info.GuildID = ID
		s.GuildID = append(s.GuildID, info)
		s.SaveConfigClient()
		s.LoadStateClient()
	}
}
