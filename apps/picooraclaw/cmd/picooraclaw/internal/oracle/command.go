package oracle

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jasperan/picooraclaw/cmd/picooraclaw/internal"
	"github.com/jasperan/picooraclaw/pkg/config"
	oracledb "github.com/jasperan/picooraclaw/pkg/oracle"
)

// NewSetupOracleCommand returns the cobra command for `picooraclaw setup-oracle`.
func NewSetupOracleCommand() *cobra.Command {
	var onnxDir, onnxFile string

	cmd := &cobra.Command{
		Use:   "setup-oracle",
		Short: "Initialize Oracle schema and load ONNX model",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(onnxDir, onnxFile)
		},
	}

	cmd.Flags().StringVar(&onnxDir, "onnx-dir", "PICO_ONNX_DIR", "Oracle directory for ONNX model")
	cmd.Flags().StringVar(&onnxFile, "onnx-file", "all_MiniLM_L12_v2.onnx", "ONNX model filename")

	return cmd
}

// NewOracleInspectCommand returns the cobra command for `picooraclaw oracle-inspect`.
func NewOracleInspectCommand() *cobra.Command {
	var limit int
	var searchQuery string

	cmd := &cobra.Command{
		Use:   "oracle-inspect [table]",
		Short: "Inspect data stored in Oracle Database",
		Long: `Inspect data stored in Oracle Database.

Tables:
  (none)        Show overview with row counts for all tables
  memories      Show stored memories with embeddings
  sessions      Show chat sessions
  transcripts   Show conversation transcripts
  state         Show key-value state entries
  notes         Show daily notes
  prompts [name]  Show system prompts (IDENTITY, SOUL, AGENT, USER, TOOLS)
  config        Show stored config entries
  meta          Show schema metadata`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(args, limit, searchQuery)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Max rows to display")
	cmd.Flags().StringVarP(&searchQuery, "search", "s", "", "Semantic search (memories only)")

	return cmd
}

func loadOracleConfig() (*config.Config, error) {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	if !cfg.Oracle.Enabled {
		return nil, fmt.Errorf("Oracle is not enabled in config. Set oracle.enabled = true first")
	}
	return cfg, nil
}

func runSetup(onnxDir, onnxFile string) error {
	cfg, err := loadOracleConfig()
	if err != nil {
		return err
	}

	fmt.Println("Setting up Oracle Database for picooraclaw...")

	conn, err := oracledb.NewConnectionManager(&cfg.Oracle)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()
	fmt.Println("Connected to Oracle Database")

	if err := oracledb.InitSchema(conn.DB()); err != nil {
		return fmt.Errorf("schema initialization failed: %w", err)
	}
	fmt.Println("Schema initialized (8 tables with PICO_ prefix)")

	var embSvc *oracledb.EmbeddingService
	if cfg.Oracle.EmbeddingProvider == "api" && cfg.Oracle.EmbeddingAPIKey != "" {
		embSvc = oracledb.NewAPIEmbeddingService(conn.DB(), cfg.Oracle.EmbeddingAPIBase, cfg.Oracle.EmbeddingAPIKey, cfg.Oracle.EmbeddingModel)
		fmt.Printf("Using API embedding provider (model: %s)\n", cfg.Oracle.EmbeddingModel)
	} else {
		embSvc = oracledb.NewEmbeddingService(conn.DB(), cfg.Oracle.ONNXModel)
		loaded, err := embSvc.CheckONNXLoaded()
		if err != nil {
			fmt.Printf("Warning: Could not check ONNX model status: %v\n", err)
		}

		if loaded {
			fmt.Printf("ONNX model '%s' already loaded\n", cfg.Oracle.ONNXModel)
		} else {
			fmt.Printf("Loading ONNX model '%s'...\n", cfg.Oracle.ONNXModel)
			if err := embSvc.LoadONNXModel(onnxDir, onnxFile); err != nil {
				fmt.Printf("ONNX model load failed: %v\n", err)
				fmt.Println("  You may need to manually load the ONNX model.")
				fmt.Println("  See: https://docs.oracle.com/en/database/oracle/oracle-database/23/vecse/")
			} else {
				fmt.Printf("ONNX model '%s' loaded\n", cfg.Oracle.ONNXModel)
			}
		}
	}

	if embSvc.TestEmbedding() {
		fmt.Printf("Embedding test passed (mode: %s)\n", embSvc.Mode())
	} else {
		fmt.Printf("Warning: Embedding test failed (mode: %s)\n", embSvc.Mode())
	}

	promptStore := oracledb.NewPromptStore(conn.DB(), cfg.Oracle.AgentID)
	workspace := cfg.WorkspacePath()
	if err := promptStore.SeedFromWorkspace(workspace); err != nil {
		fmt.Printf("Warning: Prompt seeding: %v\n", err)
	} else {
		fmt.Println("Prompts seeded from workspace")
	}

	fmt.Println("\nOracle setup complete! picooraclaw is ready to use Oracle AI Database.")
	return nil
}

func runInspect(args []string, limit int, searchQuery string) error {
	cfg, err := loadOracleConfig()
	if err != nil {
		return err
	}

	conn, err := oracledb.NewConnectionManager(&cfg.Oracle)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	db := conn.DB()

	filter := ""
	subFilter := ""
	if len(args) > 0 {
		filter = args[0]
	}
	if len(args) > 1 {
		subFilter = args[1]
	}

	switch filter {
	case "":
		inspectOverview(db, cfg.Oracle.AgentID)
	case "memories":
		inspectMemories(db, cfg.Oracle.AgentID, limit, searchQuery, &cfg.Oracle)
	case "sessions":
		inspectSessions(db, cfg.Oracle.AgentID, limit)
	case "transcripts":
		inspectTranscripts(db, cfg.Oracle.AgentID, limit)
	case "state":
		inspectState(db, cfg.Oracle.AgentID)
	case "notes":
		inspectDailyNotes(db, cfg.Oracle.AgentID, limit)
	case "prompts":
		inspectPrompts(db, cfg.Oracle.AgentID, subFilter)
	case "config":
		inspectConfig(db, cfg.Oracle.AgentID)
	case "meta":
		inspectMeta(db)
	default:
		return fmt.Errorf("unknown table: %s", filter)
	}
	return nil
}
