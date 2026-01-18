package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// ビルド時に ldflags で上書きされる変数
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func initVersionInfo() {
	// go install でビルドされた場合のフォールバック
	if info, ok := debug.ReadBuildInfo(); ok {
		// バージョン情報（go install時はモジュールバージョン）
		if Version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}

		// VCS情報から取得
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if Commit == "unknown" && setting.Value != "" {
					// 短縮ハッシュ
					if len(setting.Value) > 7 {
						Commit = setting.Value[:7]
					} else {
						Commit = setting.Value
					}
				}
			case "vcs.time":
				if BuildDate == "unknown" && setting.Value != "" {
					BuildDate = setting.Value
				}
			case "vcs.modified":
				if setting.Value == "true" && Commit != "unknown" {
					Commit += "-dirty"
				}
			}
		}
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("note-cli %s\n", Version)
		fmt.Printf("  Commit:     %s\n", Commit)
		fmt.Printf("  Built:      %s\n", BuildDate)
		fmt.Printf("  Go version: %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
