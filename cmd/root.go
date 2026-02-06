package cmd

import (
	"fmt"
	"os"

	"github.com/intiramisu/note-cli/internal/config"
	"github.com/intiramisu/note-cli/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "note-cli",
	Short:   "A lightweight CLI tool for notes and tasks",
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		noteStorage, err := newStorage()
		if err != nil {
			return err
		}

		taskManager, err := newTaskManager()
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

	// ビルド情報からバージョン情報を取得（go install対応）
	initVersionInfo()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
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
