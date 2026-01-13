.PHONY: build install clean test version

# バージョン情報
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ビルドフラグ
LDFLAGS = -ldflags "-X github.com/intiramisu/note-cli/cmd.Version=$(VERSION) \
                    -X github.com/intiramisu/note-cli/cmd.Commit=$(COMMIT) \
                    -X github.com/intiramisu/note-cli/cmd.BuildDate=$(DATE)"

# デフォルトターゲット
all: build

# ビルド
build:
	go build $(LDFLAGS) -o note-cli .

# インストール（$GOPATH/bin へ）
install:
	go install $(LDFLAGS) .

# クリーン
clean:
	rm -f note-cli

# テスト
test:
	go test ./...

# 現在のバージョン情報を表示
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"
