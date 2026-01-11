package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "note-cli",
	Short: "CLIベースのメモ管理・タスク管理ツール",
	Long: `note-cli はターミナルからメモとタスクを管理するための
軽量で高速な CLI ツールです。

コマンドライン中心のワークフローを好む開発者向けに設計されています。`,
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
	home, _ := os.UserHomeDir()
	viper.SetDefault("notes_dir", home+"/notes")
	viper.SetDefault("editor", "vim")
	viper.SetDefault("default_tags", []string{})

	viper.AutomaticEnv()
	viper.ReadInConfig()
}
