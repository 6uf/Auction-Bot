package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"

	discordgo "github.com/Liza-Developer/tempbuild"
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
	Info      string    `json:"info"`
	MessageID string    `json:"messageid"`
	Claimed   bool      `json:"claimed"`
}

type History struct {
	Bid    int64  `json:"bid"`
	Bidder string `json:"bidder"`
}

type Components struct {
	Comp []FormData `json:"components"`
	Type int64      `json:"type"`
}

type FormData struct {
	ID    string `json:"custom_id"`
	Label string `json:"label"`
	Style int64  `json:"style"`
	Value string `json:"value"`
	Type  int64  `json:"type"`
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
 
    Ver: 2.25
Made By: Liza

Commands :
             
/add-staff
/auction-create
/bid
/bin-name
/delete-auction
/remove-staff
/revert-user

`)
	Data.LoadState()
}

func main() {
	s, _ = discordgo.New("Bot " + Data.Key)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Attempting to create channel..",
					Flags:   1 << 6,
				},
			}); err != nil {
				panic(err)
			}

			var User string
			var Price int
			var Information string

			for _, data := range i.ModalSubmitData().Components {
				if data, err := data.MarshalJSON(); err == nil {
					var Info Components

					if err := json.Unmarshal(data, &Info); err != nil {
						fmt.Println(err)
					}

					switch Info.Comp[0].ID {
					case "opinion":
						User += Info.Comp[0].Value
					case "price":
						if Price, err = strconv.Atoi(Info.Comp[0].Value); err != nil {
							fmt.Println(err)
						}
					case "information":
						Information += Info.Comp[0].Value
					}
				}
			}

			var has bool = false
			var use *discordgo.Channel
			guild, err := s.Guild(i.GuildID)
			if err == nil {
				if channels, err := s.GuildChannels(guild.ID); err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: fmt.Sprintf("```%v```", err),
								},
							},
							Flags: 1 << 6,
						},
					})
				} else {
					for _, channel := range channels {
						if channel.Type == discordgo.ChannelTypeGuildCategory {
							if strings.ToLower(channel.Name) == "auctions" {
								use = channel
								has = true
								break
							}
						}
					}
				}

				if !has {
					if use, err = s.GuildChannelCreate(guild.ID, "auctions", discordgo.ChannelTypeGuildCategory); err != nil {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{
								Embeds: []*discordgo.MessageEmbed{
									{
										Author:      &discordgo.MessageEmbedAuthor{},
										Color:       000000, // Green
										Description: fmt.Sprintf("```%v```", err),
									},
								},
								Flags: 1 << 6,
							},
						})
					}
				}

				if complex, err := s.GuildChannelCreate(guild.ID, User, discordgo.ChannelTypeGuildText); err == nil {
					if _, err := s.ChannelEditComplex(complex.ID, &discordgo.ChannelEdit{
						RateLimitPerUser: 10,
						ParentID:         use.ID,
					}); err == nil {
						if message, err := s.ChannelMessageSendEmbed(complex.ID, &discordgo.MessageEmbed{
							Author:      &discordgo.MessageEmbedAuthor{},
							Color:       000000, // Green
							Description: fmt.Sprintf("`%v`\nStarting Bid: `$%v`\n\n```diff\n%v```\nHow to bid?\nUse the `/bid` command.", User, Price, Information),
						},
						); err == nil {
							Data.Data = append(Data.Data, Info{
								ChannelID: complex.ID,
								StartBid:  int64(Price),
								Name:      User,
								Info:      Information,
								MessageID: message.ID,
								Claimed:   false,
							})

							Data.SaveConfig()
							Data.LoadState()

							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Embeds: []*discordgo.MessageEmbed{
										{
											Author:      &discordgo.MessageEmbedAuthor{},
											Color:       000000, // Green
											Description: fmt.Sprintf("%v Created succesfully.", User),
										},
									},
									Flags: 1 << 6,
								},
							})

							return
						} else {
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Embeds: []*discordgo.MessageEmbed{
										{
											Author:      &discordgo.MessageEmbedAuthor{},
											Color:       000000, // Green
											Description: fmt.Sprintf("```%v```", err),
										},
									},
									Flags: 1 << 6,
								},
							})
							return
						}
					} else {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{
								Embeds: []*discordgo.MessageEmbed{
									{
										Author:      &discordgo.MessageEmbedAuthor{},
										Color:       000000, // Green
										Description: fmt.Sprintf("```%v```", err),
									},
								},
								Flags: 1 << 6,
							},
						})
						return
					}
				} else {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: fmt.Sprintf("```%v```", err),
								},
							},
							Flags: 1 << 6,
						},
					})
					return
				}
			}
		}
	})

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

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
	}

	fmt.Println("Closing the program. [Caused by error or natural ctrl+c usage]")
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
