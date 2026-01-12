package cmd

import (
	"fmt"
	"os"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/note"
	"github.com/intiramisu/note-cli/internal/task"
	"github.com/intiramisu/note-cli/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "note-cli",
	Short: "CLIベースのメモ管理・タスク管理ツール",
	Long: `note-cli はターミナルからメモとタスクを管理するための
軽量で高速な CLI ツールです。

コマンドライン中心のワークフローを好む開発者向けに設計されています。

引数なしで実行すると統合TUIが起動します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Global

		noteStorage, err := note.NewStorage(cfg.NotesDir)
		if err != nil {
			return err
		}

		taskManager, err := task.NewManager(cfg.NotesDir)
		if err != nil {
			return err
		}

		return ui.Run(noteStorage, taskManager)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "設定ファイルのパス (デフォルト: ~/.config/note-cli/config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configPath := home + "/.config/note-cli"
		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// デフォルト値を設定
	config.SetDefaults()

	viper.AutomaticEnv()
	viper.ReadInConfig()

	// 設定を読み込み
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "設定の読み込みに失敗: %v\n", err)
		os.Exit(1)
	}
}
