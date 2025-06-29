package management

// func routineRegisterNewPlayers(ctx context.Context, dg *discordgo.Session, GuildID *string, db database.Database, members map[string]*discordgo.Member) {
// 	channels, err := dg.GuildChannels(*GuildID)
// 	if err != nil {
// 		log.Fatalf("Cannot get guild channels: %v", err)
// 	}

// 	for _, member := range members {
// 		if !discordutils.IsMember(member) {
// 			// log.Printf("Member %s is not a member %+v", member.User.ID, member.Roles)
// 			continue
// 		}
// 		class := getMemberWarClass(member)
// 		if class == "" {
// 			log.Printf("Member %s has no class", member.User.ID)
// 			continue
// 		}

// 		if member.Nick == "" {
// 			log.Printf("Member %s has no nickname", member.User.ID)
// 			continue
// 		}

// 		classEmoji := getClassEmoji(class)
// 		nickWithoutEmoji := strings.Trim(gomoji.RemoveEmojis(member.Nick), " ")
// 		isAdmin := discordutils.IsMemberAdmin(member)

// 		if !isAdmin && member.Nick != classEmoji+nickWithoutEmoji {
// 			_ = dg.GuildMemberNickname(*GuildID, member.User.ID, classEmoji+nickWithoutEmoji)
// 		}

// 		player, err := types.GetPlayerByDiscordID(ctx, db, member.User.ID)
// 		if err != nil {
// 			log.Fatalf("Cannot get player: %v", err)
// 			continue
// 		}
// 		if player == nil {
// 			ign := strings.Trim(gomoji.RemoveEmojis(discordutils.GetMemberName(member)), " ")
// 			log.Printf("Registering player %s", ign)
// 			newPlayer := types.Player{
// 				ID:        primitive.NewObjectID(),
// 				DiscordID: member.User.ID,
// 				IGN:       ign,
// 				WarClass:  getMemberWarClass(member),
// 			}

// 			if shouldHaveTicket(member) {
// 				existantTicket := findTicketChannel(channels, nickWithoutEmoji)
// 				if existantTicket != nil {
// 					log.Printf("Player %s has a ticket channel", nickWithoutEmoji)
// 					newPlayer.TicketChannel = existantTicket.ID
// 					updateTicketChannel(dg, &newPlayer, class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
// 				} else {
// 					ticketChannel, _ := createTicketChannel(dg, GuildID, &(member.User.ID), class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
// 					newPlayer.TicketChannel = ticketChannel.ID
// 				}
// 			}

// 			err = types.InsertPlayer(ctx, db, &newPlayer)
// 			if err != nil {
// 				log.Fatalf("Cannot create player: %v", err)
// 				continue
// 			}
// 		} else {
// 			updatedPlayer := false
// 			if !shouldHaveTicket(member) && player.TicketChannel != "" {
// 				deleteTicketChannel(dg, player)
// 				player.TicketChannel = ""
// 				updatedPlayer = true
// 			}

// 			if shouldHaveTicket(member) && player.TicketChannel == "" {
// 				ticketChannel, _ := createTicketChannel(dg, GuildID, &(member.User.ID), class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
// 				player.TicketChannel = ticketChannel.ID
// 				updatedPlayer = true
// 			}

// 			if player.ArchivedAt != nil {
// 				updateTicketChannel(dg, player, class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
// 				player.ArchivedAt = nil
// 				updatedPlayer = true
// 			}

// 			if player.WarClass != class {
// 				updateTicketChannel(dg, player, class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
// 				player.WarClass = class
// 				updatedPlayer = true
// 			}

// 			if updatedPlayer {
// 				err = types.UpdatePlayer(ctx, db, player)
// 				if err != nil {
// 					log.Fatalf("Cannot update player: %v", err)
// 				}
// 			}
// 		}
// 	}
// }

// func routineArchiveUnavailablePlayers(ctx context.Context, dg *discordgo.Session, GuildID *string, db database.Database, members map[string]*discordgo.Member) {
// 	players, err := types.GetPlayers(ctx, db)
// 	if err != nil {
// 		log.Fatalf("Cannot get players: %v", err)
// 	}

// 	for _, player := range players {
// 		if player.ArchivedAt != nil {
// 			continue
// 		}
// 		member, err := dg.GuildMember(*GuildID, player.DiscordID, discordgo.WithRetryOnRatelimit(true))
// 		if err != nil {
// 			log.Println("Cannot get member", player.DiscordID, err)
// 			if strings.Contains(err.Error(), "Unknown Member") {
// 				processPlayerArchiving(ctx, dg, db, &player)
// 			}
// 			continue
// 		}
// 		if member == nil && members[player.DiscordID] != nil {
// 			member = members[player.DiscordID]
// 		}

// 		if member == nil {
// 			fmt.Println("Archived player due to missing member", player.IGN)
// 			processPlayerArchiving(ctx, dg, db, &player)
// 		} else {
// 			if !discordutils.IsMember(member) {
// 				fmt.Println("Archived player due to not being a member", player.IGN)
// 				processPlayerArchiving(ctx, dg, db, &player)
// 			}
// 		}
// 	}
// }

// func routineUnarchiveReturningPlayers(ctx context.Context, dg *discordgo.Session, GuildID *string, db database.Database, members map[string]*discordgo.Member) {
// 	players, err := types.GetPlayers(ctx, db)
// 	if err != nil {
// 		log.Fatalf("Cannot get players: %v", err)
// 	}

// 	for _, player := range players {
// 		if player.ArchivedAt == nil {
// 			continue
// 		}
// 		member, _ := dg.GuildMember(*GuildID, player.DiscordID)
// 		if member == nil {
// 			continue
// 		}

// 		if discordutils.IsMember(member) {
// 			processPlayerUnarchiving(ctx, dg, db, &player)
// 		}
// 	}
// }

// func routineDeleteArchivedPlayers(ctx context.Context, dg *discordgo.Session, db database.Database) {
// 	players, err := types.GetArchivedPlayers(ctx, db)
// 	if err != nil {
// 		log.Fatalf("Cannot get archived players: %v", err)
// 	}

// 	for _, player := range players {
// 		if player.ArchivedAt == nil {
// 			continue
// 		}

// 		archived := (*player.ArchivedAt)
// 		if archived.Add(24 * time.Hour).Before(time.Now()) {
// 			processPlayerDeleting(ctx, dg, db, &player)
// 		}
// 	}
// }

// func routineExportPlayersCSV(ctx context.Context, db database.Database) {
// 	players, err := types.GetActivePlayers(ctx, db)
// 	if err != nil {
// 		log.Fatalf("Cannot get players: %v", err)
// 	}

// 	csvFile, err := os.Create("players_new.csv")
// 	if err != nil {
// 		log.Fatalf("Cannot create file: %v", err)
// 	}
// 	defer csvFile.Close()

// 	_, _ = csvFile.WriteString("Name,Classe")
// 	for _, player := range players {
// 		_, _ = csvFile.WriteString("\n" + player.IGN + "," + player.WarClass)
// 	}

// 	os.Remove("static/players.csv")
// 	os.Rename("players_new.csv", "static/players.csv")

// 	log.Println("Exported players to players.csv")
// }
