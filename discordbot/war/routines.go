package war

import (
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/common"
	"nwmanager/types"
	"os"
	"strings"
	"time"
)

// warCleanupRoutine runs periodically to archive wars that have reached their scheduled time
func warCleanupRoutine(ctx *common.ModuleContext) {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	defer ticker.Stop()

	log.Printf("War cleanup routine started with interval: %v", CLEANUP_INTERVAL)

	time.Sleep(5 * time.Second)

	// Run initial CSV export
	routineExportWarsCSV(ctx, ctx.DB())

	for {
		select {
		case <-ticker.C:
			err := processWarCleanup(ctx)
			if err != nil {
				log.Printf("Error in war cleanup routine: %v", err)
			}

			// Export wars to CSV
			routineExportWarsCSV(ctx, ctx.DB())

		case <-ctx.Context.Done():
			log.Println("War cleanup routine stopped")
			return
		}
	}
}

// processWarCleanup checks for wars that need to be archived
func processWarCleanup(ctx *common.ModuleContext) error {
	// Get all active wars
	wars, err := types.GetActiveWars(ctx.Context, ctx.DB())
	if err != nil {
		return err
	}

	now := time.Now()

	for _, war := range wars {
		// Check if war time has passed
		if war.ScheduledAt != nil && war.ScheduledAt.Before(now) {
			err := CancelWar(ctx, war)
			if err != nil {
				log.Printf("Error archiving war %s: %v", war.ID.Hex(), err)
				continue
			}

			log.Printf("Successfully archived war: %s (%s vs %s)",
				war.FortName, war.FortName, war.OpponentGuild)
		}
	}

	return nil
}

// cleanupWarCSVFiles removes any existing war CSV files (attack_*.csv and defense_*.csv)
func cleanupWarCSVFiles() error {
	entries, err := os.ReadDir("static")
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to clean
		}
		return fmt.Errorf("cannot read static directory: %v", err)
	}

	cleanedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		// Check if file matches war CSV pattern (attack_*.csv or defense_*.csv)
		if strings.HasPrefix(filename, "attack_") && strings.HasSuffix(filename, ".csv") ||
			strings.HasPrefix(filename, "defense_") && strings.HasSuffix(filename, ".csv") {

			filepath := fmt.Sprintf("static/%s", filename)
			err := os.Remove(filepath)
			if err != nil {
				log.Printf("Error removing war CSV file %s: %v", filepath, err)
			} else {
				log.Printf("Cleaned up war CSV file: %s", filepath)
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d war CSV files", cleanedCount)
	}

	return nil
}

// routineExportWarsCSV exports all active wars to individual CSV files
func routineExportWarsCSV(ctx *common.ModuleContext, db database.Database) {
	// First, cleanup any existing war CSV files
	err := cleanupWarCSVFiles()
	if err != nil {
		log.Printf("Error cleaning up war CSV files: %v", err)
	}

	wars, err := types.GetActiveWars(ctx.Context, db)
	if err != nil {
		log.Printf("Cannot get active wars: %v", err)
		return
	}

	if len(wars) == 0 {
		log.Println("No active wars to export")
		return
	}

	for _, war := range wars {
		err := exportWarToCSV(ctx, war)
		if err != nil {
			log.Printf("Error exporting war %s to CSV: %v", war.ID.Hex(), err)
		}
	}

	log.Printf("Exported %d active wars to CSV files", len(wars))
}

// exportWarToCSV exports a single war's participation data to a CSV file
func exportWarToCSV(ctx *common.ModuleContext, war *types.War) error {
	if war.ScheduledAt == nil {
		return fmt.Errorf("war %s has no scheduled date", war.ID.Hex())
	}

	// Create filename: ${war.Type}_${day}_${month}_${year}.csv
	warType := strings.ToLower(string(war.Type))
	day := war.ScheduledAt.Day()
	month := int(war.ScheduledAt.Month())
	year := war.ScheduledAt.Year()

	filename := fmt.Sprintf("static/%s_%02d_%02d_%d.csv", warType, day, month, year)
	tempFilename := filename + ".tmp"

	// Create temp file
	csvFile, err := os.Create(tempFilename)
	if err != nil {
		return fmt.Errorf("cannot create temp file %s: %v", tempFilename, err)
	}
	defer csvFile.Close()

	// Write CSV header
	_, err = csvFile.WriteString("IGN,WarClass,Participation")
	if err != nil {
		return fmt.Errorf("cannot write CSV header: %v", err)
	}

	// Write participation data - only for players who answered "Sim" or "Talvez"
	for playerID, participation := range war.Participations {
		// Only export players who answered "Yes" or "Maybe"
		if participation != types.WarParticipationYes && participation != types.WarParticipationMaybe {
			continue
		}

		// Get player info
		player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), playerID)
		if err != nil || player == nil || player.ArchivedAt != nil {
			continue
		}

		warClass := player.WarClass
		if warClass == "" {
			warClass = "Sem Classe"
		}

		line := fmt.Sprintf("\n%s,%s,%s", player.IGN, warClass, participation)

		_, err = csvFile.WriteString(line)
		if err != nil {
			return fmt.Errorf("cannot write CSV line: %v", err)
		}
	}

	csvFile.Close()

	// Atomically replace the old file
	err = os.Rename(tempFilename, filename)
	if err != nil {
		os.Remove(tempFilename)
		return fmt.Errorf("cannot rename temp file: %v", err)
	}

	log.Printf("Exported war %s (%s vs %s) to %s", war.FortName, war.FortName, war.OpponentGuild, filename)
	return nil
}
