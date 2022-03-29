package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

var (
	Database Guilds
	s        *discordgo.Session
)

func init() {
	clear()
	fmt.Print(`
    __  _____________  __
   /  |/  / ___/ __/ |/ /
  / /|_/ / /___\ \/    / 
 /_/  /_/\___/___/_/|_/
 
    Ver: 2.25
Made By: Liza

Commands :
             
/add-staff
/remove-staff
/auction-create
/bid
/bin-name
/delete-auction
/revert-user
/ban
/unban

`)
	Database.LoadStateClient()
}

func main() {
	s, _ = discordgo.New("Bot " + Database.Token)
	s.AddHandler(guildCreate)
	s.AddHandler(guildDelete)
	s.AddHandler(checkData)
	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates
	s.Open()

	if cmd, err := s.ApplicationCommands(s.State.User.ID, ""); err == nil {
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
	}

	r := gin.New()
	r.GET("/")

	r.Run()
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
