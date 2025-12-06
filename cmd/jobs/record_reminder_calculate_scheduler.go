package jobs

import (
	"log"

	"github.com/AsaHero/e-wallet/internal/app"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var recordReminderCalculateSchedulerCMD = &cobra.Command{
	Use:   "record-reminder-calculate-scheduler",
	Short: "Run record reminder calculate scheduler job",
	Long:  "Take all users and create tasks to calculate their reminders to record",
	Run: func(cmd *cobra.Command, args []string) {
		godotenv.Load()

		cfg, err := config.New()
		if err != nil {
			log.Fatalln("config init", err)
		}

		recordReminderCalculateScheduler, err := app.NewRecordReminderCalculateScheduler(cfg)
		if err != nil {
			log.Fatalln("app init", err)
		}

		// run application
		if err := recordReminderCalculateScheduler.Run(); err != nil {
			log.Println("record reminder calculate scheduler run", err)
		}

		// app stops
		log.Println("record reminder calculate scheduler stopping...")
		recordReminderCalculateScheduler.Stop()
		log.Println("record reminder calculate scheduler stopped gracefully")
	},
}
