package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "auction-create",
			Description: "Enter your names information.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Delay to use.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "type",
					Description: "What does the account have?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "price",
					Description: "What is the current offer of the account?",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "information",
					Description: "BIN? Bans? Ranks?",
					Required:    true,
				},
			},
		},
		{
			Name:        "bid",
			Description: "Place a bid on an account!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "Value of your bid",
					Required:    true,
				},
			},
		},
		{
			Name:        "add-mod",
			Description: "Add moderator to the config.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Name:        "role-name",
					Description: "Role to authenticate",
					Required:    true,
				},
			},
		},
		{
			Name:        "remove-mod",
			Description: "Remove moderator from the config",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Name:        "role-name",
					Description: "Role to remove",
					Required:    true,
				},
			},
		},
		{
			Name:        "done",
			Description: "Delete a auctions channel.",
		},
		{
			Name:        "revert-user",
			Description: "Revert a user(s) bids.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Name:        "role-name",
					Description: "Role to remove",
					Required:    true,
				},
			},
		},
		{
			Name:        "bin-name",
			Description: "A command admins use to finish bidding.",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"auction-create": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if !CheckAdmin(i, s) {
					return
				} else {
					name := i.ApplicationCommandData().Options[0].StringValue()
					types := i.ApplicationCommandData().Options[1].StringValue()
					price := i.ApplicationCommandData().Options[2].StringValue()
					info := i.ApplicationCommandData().Options[3].StringValue()

					has := false
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

						complex, err := s.GuildChannelCreate(guild.ID, name, discordgo.ChannelTypeGuildText)
						if err == nil {
							_, err := s.ChannelEditComplex(complex.ID, &discordgo.ChannelEdit{
								ParentID: use.ID,
							})

							if err == nil {
								message, err := s.ChannelMessageSendEmbed(complex.ID, &discordgo.MessageEmbed{
									Author: &discordgo.MessageEmbedAuthor{},
									Color:  000000, // Green
									Description: fmt.Sprintf(``+`%v`+`
Starting Bid: $%v ~ %v

Info: %v

How to bid?
Use the `+`/bid`+` command.`, "`"+name+"`", "`"+price+"`", "`"+types+"`", "`"+info+"`"),
								})

								if err == nil {
									intV, err := strconv.Atoi(price)
									if err != nil {
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
										Data.Data = append(Data.Data, Info{
											ChannelID: complex.ID,
											StartBid:  int64(intV),
											Name:      name,
											Type:      types,
											Info:      info,
											MessageID: message.ID,
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
														Description: fmt.Sprintf("%v Created succesfully.", name),
													},
												},
												Flags: 1 << 6,
											},
										})
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
						}
					}
				}
			}()
		},
		"bid": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				var id string

				if i.Member == nil {
					id = i.User.ID
				} else {
					id = i.Member.User.ID
				}

				if len(Data.Data) != 0 {
					for e, Info := range Data.Data {
						if Info.ChannelID == i.ChannelID {
							if i.ApplicationCommandData().Options[0].IntValue() >= Info.StartBid {
								if Info.History == nil {
									Info.History = append(Info.History, History{})
								}

								if i.ApplicationCommandData().Options[0].IntValue() >= Info.History[len(Info.History)-1].Bid+5 {
									if i.ApplicationCommandData().Options[0].IntValue() < 10000 {
										Data.Data[e].History = append(Data.Data[e].History, History{
											Bid:    i.ApplicationCommandData().Options[0].IntValue(),
											Bidder: id,
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
														Description: fmt.Sprintf("<@%v> Succesfully placed your bid: $%v", id, i.ApplicationCommandData().Options[0].IntValue()),
													},
												},
												Flags: 1 << 6,
											},
										})

										s.ChannelMessageEditEmbed(Info.ChannelID, Info.MessageID, &discordgo.MessageEmbed{
											Author: &discordgo.MessageEmbedAuthor{},
											Color:  000000, // Green
											Description: fmt.Sprintf(``+`%v`+`
Current Bid: $%v ~ %v
											
Type: %v
Info: %v
											
How to bid?
Use the `+`/bid`+` command.`, "`"+Info.Name+"`", "`"+fmt.Sprintf("%v", i.ApplicationCommandData().Options[0].IntValue())+"`", "<@"+id+">", "`"+Info.Type+"`", "`"+Info.Info+"`")})
									} else {
										s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
											Type: discordgo.InteractionResponseChannelMessageWithSource,
											Data: &discordgo.InteractionResponseData{
												Embeds: []*discordgo.MessageEmbed{
													{
														Author:      &discordgo.MessageEmbedAuthor{},
														Color:       000000, // Green
														Description: "Value is to large, please dm a staff member to validate your payment.",
													},
												},
												Flags: 1 << 6,
											},
										})
									}
								} else {
									s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
										Type: discordgo.InteractionResponseChannelMessageWithSource,
										Data: &discordgo.InteractionResponseData{
											Embeds: []*discordgo.MessageEmbed{
												{
													Author:      &discordgo.MessageEmbedAuthor{},
													Color:       000000, // Green
													Description: fmt.Sprintf("Please bid higher then <@%v> +5$.", Info.History[len(Info.History)-1].Bidder),
												},
											},
											Flags: 1 << 6,
										},
									})
								}
							} else {
								s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{
										Embeds: []*discordgo.MessageEmbed{
											{
												Author:      &discordgo.MessageEmbedAuthor{},
												Color:       000000, // Green
												Description: "Please bid higher then the start amount.",
											},
										},
										Flags: 1 << 6,
									},
								})
							}
						}
					}
				}
			}()
		},
		"add-mod": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if !CheckAdmin(i, s) {
					return
				} else {
					Data.IDs = append(Data.IDs, i.ApplicationCommandData().Options[0].UserValue(s).ID)

					Data.SaveConfig()
					Data.LoadState()

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: "```Succesfully added moderator to config```",
								},
							},
							Flags: 1 << 6,
						},
					})
				}
			}()
		},
		"remove-mod": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if !CheckAdmin(i, s) {
					return
				} else {
					Data.IDs = remove(Data.IDs, i.ApplicationCommandData().Options[0].UserValue(s).ID)
					Data.SaveConfig()
					Data.LoadState()

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: "```Succesfully removed moderator from config```",
								},
							},
							Flags: 1 << 6,
						},
					})
				}
			}()
		},
		"done": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if !CheckAdmin(i, s) {
					return
				} else {
					var update []Info
					for _, data := range Data.Data {
						if data.ChannelID != i.ChannelID {
							update = append(update, data)
						}
					}

					Data.Data = update
					Data.SaveConfig()
					Data.LoadState()

					s.ChannelDelete(i.ChannelID)
				}
			}()
		},
		"revert-user": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if !CheckAdmin(i, s) {
					return
				} else {
					var id = i.ApplicationCommandData().Options[0].UserValue(s).ID
					var update []History
					for e, data := range Data.Data {
						if data.ChannelID == i.ChannelID {
							for _, info := range data.History {
								if info.Bidder != id {
									update = append(update, info)
								}
							}

							sort.Slice(update, func(i, j int) bool {
								return update[i].Bid < update[j].Bid
							})

							Data.Data[e].History = update
							Data.SaveConfig()
							Data.LoadState()
							break
						}
					}
				}
			}()
		},
		"bin-name": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if !CheckAdmin(i, s) {
					return
				} else {
					for _, data := range Data.Data {
						if data.History != nil {
							if data.ChannelID == i.ChannelID {
								s.ChannelMessageSendEmbed(data.ChannelID, &discordgo.MessageEmbed{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: fmt.Sprintf(`<@%v> Has outbidded %v people! make a ticket to claim your user.`, data.History[len(data.History)-1].Bidder, "`"+fmt.Sprintf("%v", len(data.History))+"`")})
							}
						}
					}
				}
			}()
		},
	}
)

func CheckAdmin(i *discordgo.InteractionCreate, s *discordgo.Session) bool {
	var id string

	if i.Member == nil {
		id = i.User.ID
	} else {
		id = i.Member.User.ID
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return false
	} else {
		if guild.OwnerID == id {
			return true
		} else {
			for _, client := range Data.IDs {
				if id == client {
					return true
				}
			}
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Author:      &discordgo.MessageEmbedAuthor{},
					Color:       000000, // Green
					Description: "```You are not authorized to use this Bot.```",
					Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
					Title:       "Errors",
				},
			},
			Flags: 1 << 6,
		},
	})

	return false
}

func remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			l = append(l[:i], l[i+1:]...)
		}
	}
	return l
}
