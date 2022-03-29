package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type History struct {
	Bid    int64  `json:"bid"`
	Bidder string `json:"bidder"`
}

type RoleSpecific struct {
	Role   bool   `json:"roleSP"`
	RoleID string `json:"roleid"`
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

func WriteFile(path string, content string) {
	ioutil.WriteFile(path, []byte(content), 0644)
}

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	Database.AddGuild(event.Guild.ID)
	s.UpdateListeningStatus(fmt.Sprintf("%v Servers", len(s.State.Guilds)))
}

func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	s.UpdateListeningStatus(fmt.Sprintf("%v Servers", len(s.State.Guilds)))
}

func checkData(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		var Role string

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
				case "roleid":
					Role = Info.Comp[0].Value
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
						var ROLE RoleSpecific
						if Role != "0" || len(Role) > 17 {
							ROLE = RoleSpecific{
								Role:   true,
								RoleID: Role,
							}
						} else {
							ROLE = RoleSpecific{
								Role:   false,
								RoleID: "",
							}
						}

						Data := Database.GetGuildData(i.GuildID)

						Data.Data = append(Data.Data, Info{
							ChannelID: complex.ID,
							StartBid:  int64(Price),
							Name:      User,
							Info:      Information,
							MessageID: message.ID,
							Claimed:   false,
							Roles:     ROLE,
						})

						Database.UpdateInput(Data)
						Database.SaveConfigClient()
						Database.LoadStateClient()

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
}
