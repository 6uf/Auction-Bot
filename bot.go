package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "auction-create",
			Description: "Enter your names information.",
			Type:        discordgo.ChatApplicationCommand,
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
			Name:        "add-staff",
			Description: "Add moderator to the config.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role-name",
					Description: "Role to authenticate",
					Required:    true,
				},
			},
		},
		{
			Name:        "remove-staff",
			Description: "Remove moderator from the config",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionRole,
					Name:        "role-name",
					Description: "Role to remove",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete-auction",
			Description: "Delete a auctions channel.",
		},
		{
			Name:        "revert-user",
			Description: "Revert a user(s) bids.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Name:        "user",
					Description: "User to revert",
					Required:    true,
				},
			},
		},
		{
			Name:        "bin-name",
			Description: "A command admins use to finish bidding.",
		},
		{
			Name:        "ban",
			Description: "ban a user from bidding.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Name:        "user",
					Description: "User to ban",
					Required:    true,
				},
			},
		},
		{
			Name:        "unban",
			Description: "unban a user from bidding.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionMentionable,
					Name:        "user",
					Description: "User to unban",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"auction-create": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseModal,
						Data: &discordgo.InteractionResponseData{
							CustomID: "auctions",
							Title:    "Auction Data",
							Components: []discordgo.MessageComponent{
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.TextInput{
											CustomID:    "opinion",
											Label:       "Username",
											Style:       discordgo.TextInputShort,
											Placeholder: "Username of said account.",
											Required:    true,
											MaxLength:   16,
											MinLength:   1,
										},
									},
								},
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.TextInput{
											CustomID:  "price",
											Label:     "How much will this cost?",
											Style:     discordgo.TextInputShort,
											Required:  true,
											MaxLength: 35,
										},
									},
								},
								discordgo.ActionsRow{
									Components: []discordgo.MessageComponent{
										discordgo.TextInput{
											CustomID:  "information",
											Label:     "bans? gc? tid? basic information.",
											Style:     discordgo.TextInputParagraph,
											Required:  true,
											MaxLength: 2000,
										},
									},
								},
							},
						},
					})
					if err != nil {
						panic(err)
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

							if Info.Claimed {
								s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{
										Embeds: []*discordgo.MessageEmbed{
											{
												Author:      &discordgo.MessageEmbedAuthor{},
												Color:       000000, // Green
												Description: "This Auction is already claimed and you cannot bid any further.",
											},
										},
										Flags: 1 << 6,
									},
								})
								return
							}

							if i.ApplicationCommandData().Options[0].IntValue() > Info.StartBid+5 {

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
											Author:      &discordgo.MessageEmbedAuthor{},
											Color:       000000, // Green
											Description: fmt.Sprintf("`%v`\nCurrent Bid: `$%v` ~ <@%v>\n\n```diff\n%v```\nHow to bid?\nUse the `/bid` command.", Info.Name, i.ApplicationCommandData().Options[0].IntValue(), id, Info.Info)})
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
		"add-staff": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					Role := i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID)
					Data.IDs = append(Data.IDs, Role.ID)

					Data.SaveConfig()
					Data.LoadState()

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: fmt.Sprintf("```Succesfully added %v to config```", Role.Name),
								},
							},
							Flags: 1 << 6,
						},
					})
				}
			}()
		},
		"remove-staff": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					Data.IDs = remove(Data.IDs, i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID).ID)
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
		"delete-auction": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					var update []Info
					for _, data := range Data.Data {
						if data.ChannelID != i.ChannelID {
							update = append(update, data)
						} else {
							s.ChannelDelete(i.ChannelID)
						}
					}

					Data.Data = update
					Data.SaveConfig()
					Data.LoadState()

					s.ChannelMessageSend(i.ChannelID, "Couldnt delete channel, it isnt a auction.")
				}
			}()
		},
		"revert-user": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

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

							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Embeds: []*discordgo.MessageEmbed{
										{
											Author:      &discordgo.MessageEmbedAuthor{},
											Color:       000000, // Green
											Description: fmt.Sprintf("```Succesfully removed %v Bid(s)```", i.ApplicationCommandData().Options[0].UserValue(s).Username),
										},
									},
									Flags: 1 << 6,
								},
							})

							for _, data := range Data.Data {
								if data.ChannelID == i.ChannelID {
									if len(data.History) == 0 {
										s.ChannelMessageEditEmbed(data.ChannelID, data.MessageID, &discordgo.MessageEmbed{
											Author:      &discordgo.MessageEmbedAuthor{},
											Color:       000000, // Green
											Description: fmt.Sprintf("`%v`\nStarting Bid: `$%v`\n\n```diff\n%v```\nHow to bid?\nUse the `/bid` command.", data.Name, data.StartBid, data.Info),
										})
									} else {
										if data.ChannelID == i.ChannelID {
											s.ChannelMessageEditEmbed(data.ChannelID, data.MessageID, &discordgo.MessageEmbed{
												Author:      &discordgo.MessageEmbedAuthor{},
												Color:       000000, // Green
												Description: fmt.Sprintf("`%v`\nCurrent Bid: `$%v` ~ <@%v>\n\n```diff\n%v```\nHow to bid?\nUse the `/bid` command.", data.Name, data.History[len(data.History)-1].Bid, data.History[len(data.History)-1].Bidder, data.Info)})
										}
									}
									break
								}
							}

							return
						}
					}
				}
			}()
		},
		"bin-name": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					for value, data := range Data.Data {
						if data.History != nil {
							if data.ChannelID == i.ChannelID {
								if !data.Claimed {
									s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
										Type: discordgo.InteractionResponseChannelMessageWithSource,
										Data: &discordgo.InteractionResponseData{
											Embeds: []*discordgo.MessageEmbed{
												{
													Author:      &discordgo.MessageEmbedAuthor{},
													Color:       000000, // Green
													Description: fmt.Sprintf(`<@%v> Has outbidded %v people!`, data.History[len(data.History)-1].Bidder, "`"+fmt.Sprintf("%v", len(data.History))+"`"),
												},
											},
										},
									})

									Data.Data[value].Claimed = true
									Data.SaveConfig()
									Data.LoadState()

									if roles, err := s.GuildRoles(i.GuildID); err != nil {
										fmt.Println(err)
									} else {
										for _, info := range roles {
											if err := s.ChannelPermissionSet(data.ChannelID, info.ID, discordgo.PermissionOverwriteTypeRole, discordgo.PermissionViewChannel, discordgo.PermissionSendMessages); err != nil {
												fmt.Println(err)
											}
										}
									}

									if user, err := s.User(data.History[len(data.History)-1].Bidder); err != nil {
										fmt.Println(err)
									} else {
										if channel, err := s.GuildChannelCreate(i.GuildID, user.Username, discordgo.ChannelTypeGuildText); err != nil {
											fmt.Println(err)
										} else {

											var perms []*discordgo.PermissionOverwrite = []*discordgo.PermissionOverwrite{
												{
													ID:    i.GuildID,
													Type:  discordgo.PermissionOverwriteTypeRole,
													Deny:  discordgo.PermissionViewChannel,
													Allow: discordgo.PermissionChangeNickname,
												},
												{
													ID:    user.ID,
													Type:  discordgo.PermissionOverwriteTypeMember,
													Deny:  discordgo.PermissionBanMembers,
													Allow: discordgo.PermissionViewChannel,
												},
											}

											for _, roles := range Data.IDs {
												perms = append(perms, &discordgo.PermissionOverwrite{
													ID:    roles,
													Type:  discordgo.PermissionOverwriteTypeRole,
													Deny:  discordgo.PermissionMentionEveryone,
													Allow: discordgo.PermissionViewChannel,
												})
											}

											if _, err := s.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{
												PermissionOverwrites: perms,
											}); err != nil {
												fmt.Println(err)
											}
											var content string
											for _, roles := range Data.IDs {
												content += fmt.Sprintf("<@&%v> ", roles)
											}

											s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
												Content: content,
												Embeds: []*discordgo.MessageEmbed{
													{
														Author:      &discordgo.MessageEmbedAuthor{},
														Color:       000000, // Green
														Description: fmt.Sprintf(`Welcome <@%v> an admin will be with you shortly!`, user.ID),
													},
												},
											})
										}
									}

									return
								} else {
									s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
										Type: discordgo.InteractionResponseChannelMessageWithSource,
										Data: &discordgo.InteractionResponseData{
											Embeds: []*discordgo.MessageEmbed{
												{
													Author:      &discordgo.MessageEmbedAuthor{},
													Color:       000000, // Green
													Description: "This Auction is already Claimed.",
												},
											},
											Flags: 1 << 6,
										},
									})
									return
								}
							}
						} else {
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{
									Embeds: []*discordgo.MessageEmbed{
										{
											Author:      &discordgo.MessageEmbedAuthor{},
											Color:       000000, // Green
											Description: "```Cannot bin a name with no bids!```",
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
		"ban": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				var id string

				if i.Member == nil {
					id = i.User.ID
				} else {
					id = i.Member.User.ID
				}

				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					Data.Bans = append(Data.IDs, id)

					Data.SaveConfig()
					Data.LoadState()

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: fmt.Sprintf("```Succesfully added %v to bans```", i.User.Username),
								},
							},
							Flags: 1 << 6,
						},
					})
				}
			}()
		},
		"unban": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			go func() {
				var id string

				if i.Member == nil {
					id = i.User.ID
				} else {
					id = i.Member.User.ID
				}

				if data, payload := CheckAdmin(i, s); !data {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: payload,
									Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
									Title:       "Errors",
								},
							},
							Flags: 1 << 6,
						},
					})

					return
				} else {
					Data.Bans = remove(Data.Bans, id)

					Data.SaveConfig()
					Data.LoadState()

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								{
									Author:      &discordgo.MessageEmbedAuthor{},
									Color:       000000, // Green
									Description: fmt.Sprintf("```Succesfully removed %v from bans```", i.User.Username),
								},
							},
							Flags: 1 << 6,
						},
					})
				}
			}()
		},
	}
)

func CheckAdmin(i *discordgo.InteractionCreate, s *discordgo.Session) (bool, string) {
	var id string

	if i.Member == nil {
		id = i.User.ID
	} else {
		id = i.Member.User.ID
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return false, err.Error()
	} else {
		if guild.OwnerID == id {
			return true, "Authorized"
		} else {

			for _, roles := range Data.Bans {
				if id == roles {
					return false, "You are banned from using the bot."
				}
			}

			if member, err := s.GuildMember(i.GuildID, id); err != nil {
				fmt.Println(err)
			} else {

				for _, roles := range Data.IDs {
					for _, memberRole := range member.Roles {
						if memberRole == roles {
							return true, "Authorized"
						}
					}
				}
			}
		}
	}

	return false, "You are not authorized to use this bot"
}

func remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			l = append(l[:i], l[i+1:]...)
		}
	}
	return l
}
