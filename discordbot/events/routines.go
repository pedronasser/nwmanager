package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/discordutils"
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

	EVENT_NOTIFICATION_REMINDER time.Duration = 15 * time.Minute
)

func eventsCheckRoutine(db types.Database, dg *discordgo.Session) {
	// Cleanup completed events
	ticker := time.NewTicker(EVENT_CHECK_INTERVAL)
	for {
		<-ticker.C
		fmt.Println("Checking events...")
		ctx := context.Background()
		res, err := db.Collection(types.EventsCollection).Find(ctx, bson.M{})
		if err != nil {
			log.Fatalf("Cannot get events: %v", err)
		}
		now := GetCurrentTimeAsUTC()
		for res.Next(ctx) {
			var event types.Event
			err := res.Decode(&event)
			if err != nil {
				log.Fatalf("Cannot decode event: %v", err)
			}

			if event.Status == types.EventStatusCompleted && event.CompletedAt.Add(EVENT_COMPLETE_EXPIRE_DURATION).Before(now) {
				closeEvent(ctx, db, dg, &event)
				fmt.Println("Event closed", event.ID, "completion", *event.CompletedAt, now)
				continue
			}

			if event.Status == types.EventStatusOpen && event.ScheduledAt != nil {
				notificationTime := event.ScheduledAt.Add(-1 * EVENT_NOTIFICATION_REMINDER)

				if event.NotifiedAt == nil && notificationTime.Before(now) {
					fmt.Println("Sending messages for event", event)
					go func(dg *discordgo.Session) {
						for _, player := range event.PlayerSlots {
							if player != "" {
								discordutils.SendMemberDM(dg, player, fmt.Sprintf("O evento **%s** em que você se inscreveu está agendado para iniciar em **15 minutos**. Verifique o canal de eventos para mais informações.", fmt.Sprintf("%s - %s", getEventTypeName(event.Type), event.Title)))
							}
						}
					}(dg)

					_, err := db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"notified_at": now}})
					if err != nil {
						log.Fatalf("Cannot update event: %v", err)
					}

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
