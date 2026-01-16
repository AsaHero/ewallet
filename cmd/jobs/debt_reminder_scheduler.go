package jobs

import (
	"log"

	"github.com/AsaHero/e-wallet/internal/app"
	"github.com/AsaHero/e-wallet/pkg/config"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var debtReminderSchedulerCMD = &cobra.Command{
	Use:   "debt-reminder-scheduler",
	Short: "Run debt reminder scheduler job",
	Long:  "Check for due debt reminders and send notifications to users",
	Run: func(cmd *cobra.Command, args []string) {
		godotenv.Load()

		cfg, err := config.New()
		if err != nil {
			log.Fatalln("config init", err)
		}

		debtReminderScheduler, err := app.NewDebtReminderScheduler(cfg)
		if err != nil {
			log.Fatalln("app init", err)
		}

		runNow, _ := cmd.Flags().GetBool("now")

		// run application
		if err := debtReminderScheduler.Run(runNow); err != nil {
			log.Println("debt reminder scheduler run", err)
		}

		// app stops
		log.Println("debt reminder scheduler stopping...")
		debtReminderScheduler.Stop()
		log.Println("debt reminder scheduler stopped gracefully")
	},
}

func init() {
	debtReminderSchedulerCMD.Flags().Bool("now", false, "Run debt reminder scheduler now")
}
