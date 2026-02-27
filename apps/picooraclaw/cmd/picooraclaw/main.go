// PicoOraClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoOraClaw contributors

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/agent"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/auth"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/cron"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/gateway"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/migrate"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/onboard"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/oracle"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/skills"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/status"
	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal/version"
)

func NewPicoOraClawCommand() *cobra.Command {
	short := fmt.Sprintf("%s picooraclaw - Personal AI Assistant v%s\n\n", internal.Logo, internal.GetVersion())

	cmd := &cobra.Command{
		Use:     "picooraclaw",
		Short:   short,
		Example: "picooraclaw agent",
	}

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agent.NewAgentCommand(),
		auth.NewAuthCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		migrate.NewMigrateCommand(),
		skills.NewSkillsCommand(),
		version.NewVersionCommand(),
		oracle.NewSetupOracleCommand(),
		oracle.NewOracleInspectCommand(),
		oracle.NewSeedDemoCommand(),
	)

	return cmd
}

func main() {
	cmd := NewPicoOraClawCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
