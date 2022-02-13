package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/bwmarrin/discordgo"
)

type Channels struct {
	Data []Info
	Key  string   `json:"BotKEY"`
	IDs  []string `json:"verified"`
}

type Info struct {
	ChannelID string    `json:"channelID"`
	History   []History `json:"bidhistory"`
	StartBid  int64     `json:"startbid"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Info      string    `json:"info"`
	MessageID string    `json:"messageid"`
}

type History struct {
	Bid    int64  `json:"bid"`
	Bidder string `json:"bidder"`
}

var (
	Data Channels
	s    *discordgo.Session
)

func init() {
	clear()
	fmt.Print(`
    __  _____________  __
   /  |/  / ___/ __/ |/ /
  / /|_/ / /___\ \/    / 
 /_/  /_/\___/___/_/|_/
 
    Ver: 1.0
Made By: Liza

`)
	Data.LoadState()
}

func main() {
	s, _ = discordgo.New("Bot " + Data.Key)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	s.Open()

	cmd, err := s.ApplicationCommands(s.State.User.ID, "")
	if err == nil {
		if len(cmd) == 0 {
			for _, command := range commands {
				if _, err := s.ApplicationCommandCreate(s.State.User.ID, "", command); err != nil {
					fmt.Println(err)
				}
			}
		} else {
			if _, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands); err != nil {
				fmt.Println(err)
			}
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
	} else {
		fmt.Println(err)
	}
}

func clear() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
