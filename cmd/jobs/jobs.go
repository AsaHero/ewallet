package jobs

import "github.com/spf13/cobra"

var JobsCMD = &cobra.Command{
	Use:     "jobs [command]",
	Short:   "Run jobs",
	Example: `ewallet jobs record-reminder-calculate`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	JobsCMD.AddCommand(
		recordReminderCalculateSchedulerCMD,
	)
}
