package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := viper.AllSettings()

		data, err := yaml.Marshal(settings)
		if err != nil {
			return fmt.Errorf("設定の表示に失敗: %w", err)
		}

		if configFile := viper.ConfigFileUsed(); configFile != "" {
			fmt.Printf("設定ファイル: %s\n", configFile)
		} else {
			fmt.Println("設定ファイル: (未作成・デフォルト値を使用)")
		}
		fmt.Println("---")
		fmt.Print(string(data))

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// ハイフン形式のキーをアンダースコアに変換
		key = strings.ReplaceAll(key, "-", "_")

		viper.Set(key, value)

		// 設定ファイルのパスを決定
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
			}
			configFile = filepath.Join(home, ".config", "note-cli", "config.yaml")
		}

		// ディレクトリがなければ作成
		configDir := filepath.Dir(configFile)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("設定ディレクトリの作成に失敗: %w", err)
		}

		if err := viper.WriteConfigAs(configFile); err != nil {
			return fmt.Errorf("設定の保存に失敗: %w", err)
		}

		fmt.Printf("%s = %s\n", key, value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}
