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
# 統合TUIを起動（メモ一覧 → 詳細+タスク）
note-cli

# メモを作成（エディタが開く）
note-cli create "買い物リスト"

# タスク管理TUIを開く
note-cli t
```

## ショートカット

よく使うコマンドは短く書けます:

```bash
# ルート直下のショートカット
note-cli create "メモ"     # = note-cli note create
note-cli list              # = note-cli note list
note-cli show "メモ"       # = note-cli note show
note-cli edit "メモ"       # = note-cli note edit
note-cli search "検索"     # = note-cli note search

# エイリアス
note-cli n create "メモ"   # n = note
note-cli t                 # t = task
```

## メモ機能

### メモを作成

```bash
# 新規メモを作成してエディタで開く
note-cli create "会議メモ"

# タグ付きで作成
note-cli create "Goの勉強" -t go -t programming
```

### メモ一覧

```bash
# すべてのメモを表示
note-cli list

# タグでフィルタ
note-cli list --tag go
```

### メモを表示・編集・削除

```bash
# メモの内容を表示
note-cli show "会議メモ"

# メモを編集（エディタで開く）
note-cli edit "会議メモ"

# メモを削除
note-cli n delete "会議メモ"

# 確認なしで削除
note-cli n delete "会議メモ" -f
```

### メモを検索

```bash
# 全文検索
note-cli search "TODO"
```

### デイリーノート

```bash
# 今日のデイリーノートを開く（なければ作成）
note-cli d

# 昨日・明日
note-cli d yesterday
note-cli d tomorrow

# 日付指定
note-cli d 2025-01-11
note-cli d -1    # 1日前
note-cli d +3    # 3日後
```

デイリーノートは `~/notes/daily/` に保存されます。

### テンプレート

```bash
# テンプレートを使ってメモ作成
note-cli create "週次MTG" -T meeting
```

テンプレートは `~/notes/.templates/` に配置します:

```
.templates/
├── meeting.md
├── daily.md     # デイリーノートで使用
└── review.md
```

テンプレート内で使える変数:
- `{{title}}` - メモタイトル
- `{{date}}` - 日付 (デイリーノート用)
- `{{year}}`, `{{month}}`, `{{day}}`, `{{weekday}}`

## タスク機能

### TUI モード（おすすめ）

```bash
# 引数なしで実行するとTUIが起動
note-cli t
```

**TUI操作方法:**

| キー | 操作 |
|------|------|
| `h` / `←` | 左のセクションへ |
| `l` / `→` | 右のセクションへ |
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
note-cli t add "牛乳を買う"

# 優先度付きで追加（1:高, 2:中, 3:低）
note-cli t add "レポート提出" -p 1

# メモに紐づけて追加
note-cli t add "議事録まとめ" -n "会議メモ"

# タスク一覧（紐づきメモも表示）
note-cli t list

# 完了済みも含めて表示
note-cli t list -a

# タスクを完了
note-cli t done 1

# タスクを削除
note-cli t delete 1
```

## 統合TUI（メモ+タスク連携）

引数なしで `note-cli` を実行すると、メモとタスクを連携管理できる統合TUIが起動します。

```bash
note-cli
```

**操作方法:**

| キー | 操作 |
|------|------|
| `j` / `k` | 上下移動 |
| `Enter` | メモ詳細+関連タスク表示 |
| `Tab` / `Esc` | メモ一覧に戻る |
| `i` | タスク追加（自動でメモに紐づけ） |
| `d` | タスク削除 |
| `o` | タスクの紐づけ解除 |
| `Space` | タスク完了/未完了切替 |
| `q` | 終了 |

メモを選んでEnterを押すと、そのメモの内容と関連タスクが表示されます。
ここでタスクを追加すると、自動的にそのメモに紐づきます。

## 設定

デフォルトの設定ファイル: `~/.config/note-cli/config.yaml`

```bash
# サンプル設定ファイルをコピーして使用
mkdir -p ~/.config/note-cli
cp config.yaml.example ~/.config/note-cli/config.yaml
```

### 基本設定

```yaml
# メモの保存先ディレクトリ
notes_dir: ~/notes

# 使用するエディタ
editor: vim

# デフォルトタグ
default_tags: []
```

### パス設定

`notes_dir` からの相対パスで指定します。

```yaml
paths:
  templates_dir: .templates   # テンプレートディレクトリ
  tasks_file: .tasks.yaml     # タスク保存ファイル
  daily_dir: daily            # デイリーノートディレクトリ
```

### 日付フォーマット

Go の日付フォーマット形式で指定します。

```yaml
formats:
  date: "2006-01-02"          # デイリーノート名、メモ一覧
  datetime: "2006-01-02 15:04" # メモ詳細の作成日・更新日
```

### テーマ設定

カラーは hex (`#RRGGBB`) または 256色番号 (`0-255`) で指定できます。

```yaml
theme:
  colors:
    title: "#cd7cf4"          # タイトル・見出し
    selected: "#d75fd7"       # 選択中のアイテム
    done: "#626262"           # 完了済みタスク
    priority_high: "#ff0000"  # P1
    priority_medium: "#ffaf00" # P2
    priority_low: "#5fafff"   # P3

  symbols:
    cursor: "▸ "              # カーソル（選択中）
    checkbox_empty: "[ ]"     # 未完了
    checkbox_done: "[✓]"      # 完了
    note_icon: "📄"
    task_icon: "📋"
    daily_icon: "📅"

  sections:                   # タスクTUIのセクション名
    p1: "🔥 P1"
    p2: "⚡ P2"
    p3: "📝 P3"
    done: "✅ 完了"
```

### 表示設定

```yaml
display:
  separator_width: 40         # 区切り線の幅
  search_truncate: 80         # 検索結果の行切り詰め幅
  task_char_limit: 100        # タスク説明の最大文字数
  input_width: 40             # 入力フィールドの幅
```

詳細は `config.yaml.example` を参照してください。

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
