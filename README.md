# memoya

memoya は Todo管理とメモ機能を提供するMCP (Model Context Protocol) サーバーです。AIとの対話を通じてタスクやアイデアを管理できます。

## アーキテクチャ

memoyaは2つのデプロイメント方式をサポートしています：

### 1. ローカル実行（従来方式）
```
Claude Desktop ↔ MCP Client (Local) ↔ Firestore
```

### 2. Cloud Run対応（新方式）
```
Claude Desktop ↔ MCP Client (Local) ↔ HTTP ↔ Cloud Run Server ↔ Firestore
```

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

### タグ管理
- **一覧表示**: 全ての一意なタグを表示
- **自動集計**: Todo・メモ横断でのタグ分析

## セットアップ

### 必要な環境
- Go 1.22以上
- Firebase プロジェクト（Firestore使用）
- Cloud Run（クラウド実行の場合）

### ローカル実行の場合

#### インストール

```bash
# ソースからインストール
git clone https://github.com/pankona/memoya.git
cd memoya
make install

# または go install を直接使用
go install github.com/pankona/memoya/cmd/memoya@latest
```

#### Firebase設定

1. [Firebase Console](https://console.firebase.google.com/)でプロジェクトを作成
2. Firestore Databaseを有効化（テストモードでOK）
3. プロジェクト設定 → サービスアカウント → 新しい秘密鍵の生成
4. ダウンロードしたJSONファイルを安全な場所に保存

#### 環境変数の設定

```bash
# .envファイルを作成（推奨）
cat > .env << EOF
FIREBASE_PROJECT_ID=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=./path-to-service-account-key.json
EOF
```

#### Claude Desktop設定（ローカル実行）

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

### Cloud Run実行の場合

#### 1. Cloud Run Serverのデプロイ

```bash
# プロジェクト準備
git clone https://github.com/pankona/memoya.git
cd memoya

# 依存関係取得
go mod tidy

# OpenAPIからコード生成
make generate

# Docker imageビルド
make docker-build

# Cloud Runにデプロイ
gcloud run deploy memoya-server \
  --image memoya-server \
  --platform managed \
  --region us-central1 \
  --set-env-vars PROJECT_ID=your-firebase-project-id
```

#### 2. MCP Clientの設定

```bash
# MCP clientをビルド・インストール
make build
make install
```

#### 3. Claude Desktop設定（Cloud Run）

```json
{
  "mcpServers": {
    "memoya": {
      "command": "memoya",
      "env": {
        "MEMOYA_CLOUD_RUN_URL": "https://memoya-server-xxxxx-uc.a.run.app",
        "MEMOYA_AUTH_TOKEN": "your-jwt-token"
      }
    }
  }
}
```

## 開発

### 開発用コマンド

```bash
# 依存関係取得
go mod tidy

# OpenAPIからコード生成
make generate

# コードフォーマット
make fmt

# 静的解析
make lint

# テスト実行
make test

# ローカルサーバー起動（開発用）
make run-server

# MCPクライアント起動（開発用）
make run-client
```

### プロジェクト構造

```
memoya/
├── api/                    # OpenAPI仕様
│   ├── openapi.yaml       # API仕様書
│   ├── server-config.yaml # サーバー生成設定
│   └── client-config.yaml # クライアント生成設定
├── cmd/
│   ├── memoya/            # MCP Client
│   └── memoya-server/     # Cloud Run Server
├── internal/
│   ├── client/            # HTTP client & MCP bridge
│   ├── generated/         # OpenAPI生成コード
│   ├── handlers/          # ビジネスロジック
│   ├── server/            # HTTP server実装
│   ├── storage/           # Firestore抽象化
│   └── models/            # データモデル
├── Dockerfile             # Cloud Run用
├── Makefile              # ビルドコマンド
└── CLAUDE.md             # 開発者向け詳細仕様
```

### API仕様

REST APIの詳細は[OpenAPI仕様書](./api/openapi.yaml)を参照してください。

主要エンドポイント：
- `POST /mcp/memo_create` - メモ作成
- `POST /mcp/memo_list` - メモ一覧
- `POST /mcp/todo_create` - Todo作成  
- `POST /mcp/todo_list` - Todo一覧
- `POST /mcp/search` - 統合検索
- `POST /mcp/tag_list` - タグ一覧

## 利用可能なツール

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

#### 検索・分析
- `search`: Todo/メモの横断検索
- `tag_list`: 全ての一意なタグを表示

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

あなた: 「使用されているすべてのタグを表示して」

Claude: 現在使用されているタグは以下の5つです：
- work (使用頻度: 3)
- personal (使用頻度: 2)  
- urgent (使用頻度: 1)
- notes (使用頻度: 2)
- ideas (使用頻度: 1)
```

## トラブルシューティング

### ローカル実行の問題

#### Firebaseに接続できない
- 環境変数が正しく設定されているか確認
- サービスアカウントキーのパスが絶対パスか確認
- Firebaseプロジェクトでアクセス権限があるか確認

#### .envファイルが読み込まれない
- memoyaを実行するディレクトリに.envファイルがあるか確認
- ファイルの権限を確認 (`chmod 600 .env`)

### Cloud Run実行の問題

#### MCP ClientがCloud Runに接続できない
- `MEMOYA_CLOUD_RUN_URL`が正しく設定されているか確認
- Cloud Runサービスが起動しているか確認
- ファイアウォール設定を確認

#### 認証エラー
- `MEMOYA_AUTH_TOKEN`が有効か確認
- Cloud Run側の認証設定を確認

### Claude Desktop共通の問題

#### MCPツールが認識されない
- claude_desktop_config.jsonの構文エラーがないか確認
- memoyaコマンドがPATHに含まれているか確認 (`which memoya`)
- Claude Desktopを完全に再起動（タスクトレイからも終了）

#### 性能が遅い
- Cloud Run: リージョン設定とコールドスタート対策
- ローカル: Firestore接続の最適化

## ライセンス

MIT License

## 貢献

プルリクエストや Issue の作成を歓迎します。

詳細な開発ガイドラインは[CLAUDE.md](./CLAUDE.md)を参照してください。

## 今後の改善予定

### 近期
- Web UI実装（Cloud Run + SPA）
- 認証機能の強化（Google OAuth 2.0）
- CI/CD パイプライン

### 中期  
- 多言語対応
- リアルタイム同期
- より詳細な検索機能
- バルク操作のサポート

### 長期
- マルチユーザー対応
- プラグインシステム
- AIによる自動タグ付け
- 高度な分析・レポート機能