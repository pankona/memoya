# memoya

memoya は Todo管理とメモ機能を提供するMCP (Model Context Protocol) サーバーです。AIとの対話を通じてタスクやアイデアを管理できます。

## 機能

### Todo管理
- **作成・更新・削除**: Todoアイテムの基本的なCRUD操作
- **ステータス管理**: backlog, todo, in_progress, done の4段階
- **優先度設定**: high/normal の優先度レベル
- **階層構造**: 親子関係を持つTodoの管理
- **タグ機能**: 複数のタグによる分類
- **タイムスタンプ**: 作成日時、最終更新日時、完了日時の自動記録

### メモ機能
- **作成・更新・削除**: メモの基本的なCRUD操作
- **Todoとの連携**: メモからTodoへのリンク機能
- **タグ機能**: 複数のタグによる分類
- **タイムスタンプ**: 作成日時、最終更新日時の自動記録

### 検索機能
- **キーワード検索**: タイトルと説明文での全文検索
- **タグフィルタ**: 特定のタグでの絞り込み
- **タイプフィルタ**: Todo/メモ/全体での検索

## セットアップ

### 必要な環境
- Go 1.22以上
- Firebase プロジェクト（Firestore使用）

### インストール

```bash
# ソースからインストール
git clone https://github.com/pankona/memoya.git
cd memoya
make install

# または go install を直接使用
go install github.com/pankona/memoya/cmd/memoya@latest
```

### Firebase設定

1. [Firebase Console](https://console.firebase.google.com/)でプロジェクトを作成
2. Firestore Databaseを有効化（テストモードでOK）
3. プロジェクト設定 → サービスアカウント → 新しい秘密鍵の生成
4. ダウンロードしたJSONファイルを安全な場所に保存

### 環境変数の設定

```bash
# .envファイルを作成（推奨）
cat > .env << EOF
FIREBASE_PROJECT_ID=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=./path-to-service-account-key.json
EOF

# または環境変数として設定
export FIREBASE_PROJECT_ID=your-firebase-project-id
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
```

## 使用方法

### Claude Desktopでの利用

1. Claude Desktopの設定ファイルを開く
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. 以下の設定を追加:

```json
{
  "mcpServers": {
    "memoya": {
      "command": "memoya",
      "env": {
        "FIREBASE_PROJECT_ID": "your-project-id",
        "GOOGLE_APPLICATION_CREDENTIALS": "/absolute/path/to/service-account-key.json"
      }
    }
  }
}
```

3. Claude Desktopを再起動

### コマンドラインでの起動
```bash
# 環境変数を設定して起動
memoya

# または .env ファイルがある場合
memoya  # .envファイルは自動的に読み込まれます
```

### 利用可能なツール

#### Todo操作
- `todo_create`: 新しいTodoを作成
- `todo_list`: Todoリストを取得（フィルタ機能付き）
- `todo_update`: 既存のTodoを更新
- `todo_delete`: Todoを削除

#### メモ操作
- `memo_create`: 新しいメモを作成
- `memo_list`: メモリストを取得（フィルタ機能付き）
- `memo_update`: 既存のメモを更新
- `memo_delete`: メモを削除

#### 検索
- `search`: Todo/メモの横断検索

### 使用例

Claude Desktopで以下のような対話が可能です：

```
あなた: 「memoyaで新しいTodoを作成して。タイトルは'プロジェクトの企画書を作成'で、優先度は高にして」

Claude: todo_createツールを使用してTodoを作成しました：
- タイトル: プロジェクトの企画書を作成
- ステータス: todo
- 優先度: high
- ID: abc123

あなた: 「workタグが付いているTodoをすべて表示して」

Claude: workタグが付いているTodoは以下の3件です：
1. プロジェクトの企画書を作成 (優先度: high, ステータス: todo)
2. チームミーティング準備 (優先度: normal, ステータス: in_progress)
3. 月次レポート提出 (優先度: high, ステータス: done)
```

## 開発

### コードフォーマット
```bash
make fmt
```

### 静的解析
```bash
make lint
```

### テスト
```bash
make test
```

## トラブルシューティング

### Firebaseに接続できない
- 環境変数が正しく設定されているか確認
- サービスアカウントキーのパスが絶対パスか確認
- Firebaseプロジェクトでアクセス権限があるか確認

### .envファイルが読み込まれない
- memoyaを実行するディレクトリに.envファイルがあるか確認
- ファイルの権限を確認 (`chmod 600 .env`)

### Claude Desktopで使えない
- claude_desktop_config.jsonの構文エラーがないか確認
- memoyaコマンドがPATHに含まれているか確認 (`which memoya`)
- Claude Desktopを完全に再起動（タスクトレイからも終了）

## ライセンス

MIT License

## 貢献

プルリクエストや Issue の作成を歓迎します。

## 今後の改善予定

- Cloud Functions化によるセットアップの簡略化
- インメモリストレージの実装（開発用）
- より詳細な検索機能
- バルク操作のサポート
- テストケースの追加
- パフォーマンスの最適化

詳細は[TODO.md](./TODO.md)を参照してください。