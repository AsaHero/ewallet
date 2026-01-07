package jobs

import (
	"log"

	"github.com/AsaHero/e-wallet/internal/app"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var recordReminderSchedulerCMD = &cobra.Command{
	Use:   "record-reminder-scheduler",
	Short: "Run record reminder scheduler job",
	Long:  "Take all users and create tasks to schedule their reminders to record",
	Run: func(cmd *cobra.Command, args []string) {
		godotenv.Load()

		cfg, err := config.New()
		if err != nil {
			log.Fatalln("config init", err)
		}

		recordReminderScheduler, err := app.NewRecordReminderScheduler(cfg)
		if err != nil {
			log.Fatalln("app init", err)
		}

		// run application
		if err := recordReminderScheduler.Run(); err != nil {
			log.Println("record reminder scheduler run", err)
		}

		// app stops
		log.Println("record reminder scheduler stopping...")
		recordReminderScheduler.Stop()
		log.Println("record reminder scheduler stopped gracefully")
	},
}
