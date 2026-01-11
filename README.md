# note-cli

ターミナルからメモとタスクを管理する軽量な CLI ツール。

## インストール

```bash
go install github.com/intiramisu/note-cli@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/intiramisu/note-cli.git
cd note-cli
go build -o note-cli .
```

## クイックスタート

```bash
# メモを作成（エディタが開く）
note-cli note create "買い物リスト"

# タスク管理をTUIで開く
note-cli task
```

## メモ機能

### メモを作成

```bash
# 新規メモを作成してエディタで開く
note-cli note create "会議メモ"

# タグ付きで作成
note-cli note create "Goの勉強" -t go -t programming
```

### メモ一覧

```bash
# すべてのメモを表示
note-cli note list

# タグでフィルタ
note-cli note list --tag go
```

### メモを表示・編集・削除

```bash
# メモの内容を表示
note-cli note show "会議メモ"

# メモを編集（エディタで開く）
note-cli note edit "会議メモ"

# メモを削除
note-cli note delete "会議メモ"

# 確認なしで削除
note-cli note delete "会議メモ" -f
```

### メモを検索

```bash
# 全文検索
note-cli note search "TODO"
```

## タスク機能

### TUI モード（おすすめ）

```bash
# 引数なしで実行するとTUIが起動
note-cli task
```

**TUI操作方法:**

| キー | 操作 |
|------|------|
| `j` / `↓` | 下に移動 |
| `k` / `↑` | 上に移動 |
| `Enter` / `Space` | 完了/未完了を切替 |
| `i` | 新規タスク追加 |
| `d` / `x` | タスクを削除 |
| `q` | 終了 |

**タスク追加時:**

| キー | 操作 |
|------|------|
| `Tab` | 優先度を変更 (P1 → P2 → P3 → P1) |
| `Shift+Tab` | 優先度を逆順で変更 (P1 ← P2 ← P3 ← P1) |
| `Enter` | 確定 |
| `Esc` | キャンセル |

タスクは優先度ごとにセクション分けして表示されます。ターミナルのサイズに合わせてレイアウトが自動調整されます。

### CLI モード

```bash
# タスクを追加
note-cli task add "牛乳を買う"

# 優先度付きで追加（1:高, 2:中, 3:低）
note-cli task add "レポート提出" -p 1

# タスク一覧
note-cli task list

# 完了済みも含めて表示
note-cli task list -a

# タスクを完了
note-cli task done 1

# タスクを削除
note-cli task delete 1
```

## 設定

デフォルトの設定ファイル: `~/.config/note-cli/config.yaml`

```yaml
# メモの保存先ディレクトリ
notes_dir: ~/notes

# 使用するエディタ
editor: vim

# デフォルトタグ
default_tags: []
```

### 設定の確認

```bash
note-cli --help
```

## データ形式

### メモ

メモは Markdown 形式で保存されます:

```markdown
---
title: メモタイトル
created: 2025-01-11T10:30:00+09:00
modified: 2025-01-11T10:30:00+09:00
tags: [go, cli]
---

# メモタイトル

メモの内容...
```

### タスク

タスクは `~/notes/.tasks.yaml` に保存されます。

## ライセンス

MIT
