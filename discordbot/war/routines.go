package war

import (
	"context"
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	. "nwmanager/helpers"
	"nwmanager/types"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	EVENT_CHECK_INTERVAL           time.Duration = 30 * time.Second
	EVENT_COMPLETE_EXPIRE_DURATION time.Duration = 15 * time.Minute
	EVENT_MAX_DURATION             time.Duration = 60 * time.Minute

	EVENT_CONFIRMATION_REQUEST_REMINDER time.Duration = 6 * time.Hour
	EVENT_NOTIFICATION_REMINDER         time.Duration = 15 * time.Minute
)

func eventsCheckRoutine(db database.Database, dg *discordgo.Session, guildID string) {
	// Cleanup completed events
	ticker := time.NewTicker(EVENT_CHECK_INTERVAL)
	for {
		<-ticker.C
		fmt.Println("Checking wars...")
		ctx := context.Background()
		res, err := db.Collection(globals.DB_PREFIX+types.WarsCollection).Find(ctx, bson.M{})
		if err != nil {
			log.Fatalf("Cannot get wars: %v", err)
		}
		now := GetCurrentTimeAsUTC()
		for res.Next(ctx) {
			var war types.War
			err := res.Decode(&war)
			if err != nil {
				log.Fatalf("Cannot decode event: %v", err)
			}

			if war.Status == types.EventStatusOpen && war.ScheduledAt != nil {
				notificationTime := war.ScheduledAt.Add(-1 * EVENT_NOTIFICATION_REMINDER)

				if war.NotifiedAt == nil && notificationTime.Before(now) {
					sendReminderNotification(ctx, db, dg, &war)

					continue
				}

				if war.ConfirmedAt == nil && war.ScheduledAt.Add(EVENT_CONFIRMATION_REQUEST_REMINDER).Before(now) {
					members, err := discordutils.GetGuildMembers(dg, guildID, MEMBER_ROLE_ID)
					if err != nil {
						log.Fatalf("Cannot get guild members: %v", err)
					}
					sendConfirmationRequest(ctx, db, dg, &war, members)
					continue
				}

				// if event.ScheduledAt.Add(EVENT_COMPLETE_EXPIRE_DURATION).Before(now) {
				// 	closeEvent(ctx, db, dg, &event)
				// 	fmt.Println("Event closed", event.ID, "schedule date", *event.ScheduledAt, now)
				// 	continue
				// }
			}

		}
	}
}
