# CLAUDE.md

## プロジェクト概要

**memoya**は、メモとTODOを管理するMCP（Model Context Protocol）サーバーです。
Claude Desktopと統合して、AIとの会話の中でメモやタスクを効率的に管理できます。

### 主な機能
- **メモ管理**: タイトル、説明、タグ付きメモの作成・更新・削除・一覧表示
- **TODO管理**: 階層構造対応のタスク管理（ステータス、優先度、タグ）
- **統合検索**: メモとTODO横断での全文検索・タグ検索
- **タグ管理**: 全体のタグ一覧表示・分析

## アーキテクチャ

### 2つのデプロイメント方式

#### 1. ローカル実行（従来方式）
```
Claude Desktop ↔ MCP Client (Local) ↔ Firestore
```

#### 2. Cloud Run対応（新方式）
```
Claude Desktop ↔ MCP Client (Local) ↔ HTTP ↔ Cloud Run Server ↔ Firestore
```

### ディレクトリ構造
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
│   ├── handlers/          # MCPハンドラー実装
│   │   ├── memo.go        # メモ操作
│   │   ├── todo.go        # TODO操作  
│   │   ├── search.go      # 検索機能
│   │   ├── tag.go         # タグ管理
│   │   ├── *_test.go      # 単体テスト
│   │   └── mock_storage.go # テスト用モック
│   ├── server/            # HTTP server実装
│   ├── models/            # データモデル定義
│   ├── storage/           # ストレージインターフェース
│   └── config/            # 設定管理
├── Dockerfile             # Cloud Run用
├── Makefile              # ビルド・テストコマンド
└── go.mod                # Go依存関係
```

### データモデル

#### Memo（メモ）
```go
type Memo struct {
    ID           string    // 一意ID
    Title        string    // タイトル
    Description  string    // 本文
    Tags         []string  // タグ配列
    LinkedTodos  []string  // 関連TODO ID配列
    CreatedAt    time.Time // 作成日時
    LastModified time.Time // 更新日時
    ClosedAt     *time.Time // クローズ日時（オプション）
}
```

#### Todo（タスク）
```go
type Todo struct {
    ID           string       // 一意ID
    Title        string       // タイトル
    Description  string       // 詳細
    Status       TodoStatus   // ステータス（backlog/todo/in_progress/done）
    Priority     TodoPriority // 優先度（high/normal）
    Tags         []string     // タグ配列
    ParentID     string       // 親TODO ID（階層構造用）
    CreatedAt    time.Time    // 作成日時
    LastModified time.Time    // 更新日時
    ClosedAt     *time.Time   // 完了日時
}
```

### MCPツール一覧

#### メモ操作
- `memo_create` - 新規メモ作成
- `memo_list` - メモ一覧取得（タグフィルター対応）
- `memo_update` - メモ更新
- `memo_delete` - メモ削除

#### TODO操作  
- `todo_create` - 新規TODO作成
- `todo_list` - TODO一覧取得（ステータス、優先度、タグフィルター対応）
- `todo_update` - TODO更新（ステータス変更時の自動日時設定）
- `todo_delete` - TODO削除

#### 検索・分析
- `search` - 統合検索（メモ・TODO横断、クエリ・タグ・タイプフィルター）
- `tag_list` - 全タグ一覧取得

### ストレージ

#### Firestore構成
```
/memos/{memoId}     # メモコレクション
/todos/{todoId}     # TODOコレクション
```

#### フィルタリング機能
- **Todo**: ステータス、優先度、タグによる絞り込み
- **Memo**: タグによる絞り込み  
- **Search**: クエリ文字列、タグ、タイプ（memo/todo/all）による検索

### レスポンス形式

全てのMCPツールは構造化されたJSONデータをTextContentとして返します：

```json
{
  "success": true,
  "data": {...},
  "message": "操作結果メッセージ"
}
```

## 開発ルール

### コード整形
実装が一段落するたびに以下のコマンドを実行してください：
```bash
make fmt
```
または
```bash
goimports -w .
```

### 品質チェック
定期的に以下のコマンドを実行してください：
- `make lint` - 静的解析
- `make test` - テスト実行

### テスト戦略

#### 単体テスト
全ハンドラーで正常系テストを実装済み：
- `internal/handlers/*_test.go` - 各ハンドラーのCRUD操作テスト
- `internal/handlers/mock_storage.go` - テスト用ストレージモック
- テストデータによる input/output 形式検証

#### テスト実行
```bash
make test                    # 全テスト実行
go test ./internal/handlers  # ハンドラーのみテスト
go test -v ./...            # 詳細表示付き全テスト
```

### 環境設定

#### 必要な環境変数
```bash
# Firestore設定
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"

# プロジェクト設定  
export PROJECT_ID="your-firebase-project-id"
```

#### Firebase/Firestore設定
1. Firebase Console でプロジェクト作成
2. Firestore Database を作成
3. Service Account キーを生成・ダウンロード
4. `GOOGLE_APPLICATION_CREDENTIALS` で認証ファイルを指定

### ビルド・デプロイ

#### ローカルビルド
```bash
make build        # バイナリビルド
make install      # $GOPATH/bin にインストール
make clean        # ビルド成果物削除
```

#### MCP設定（Claude Desktop）
`~/.claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "memoya": {
      "command": "/path/to/memoya",
      "args": []
    }
  }
}
```

### 実装パターン

#### ハンドラー実装の基本パターン
```go
func (h *Handler) Action(ctx context.Context, ss *mcp.ServerSession, 
    params *mcp.CallToolParamsFor[Args]) (*mcp.CallToolResultFor[Result], error) {
    
    // 1. 引数取得
    args := params.Arguments
    
    // 2. バリデーション
    if h.storage == nil {
        return nil, fmt.Errorf("storage not initialized")
    }
    
    // 3. ビジネスロジック実行
    data, err := h.storage.SomeOperation(ctx, args)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    // 4. レスポンス構築
    result := Result{
        Success: true,
        Data:    data,
        Message: "operation successful",
    }
    
    // 5. JSON化してTextContentで返却
    jsonBytes, err := json.Marshal(result)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal result: %w", err)
    }
    
    return &mcp.CallToolResultFor[Result]{
        Content: []mcp.Content{
            &mcp.TextContent{Text: string(jsonBytes)},
        },
    }, nil
}
```

#### エラーハンドリング
- ストレージエラーは適切にラップして返す
- バリデーションエラーは分かりやすいメッセージで返す
- パニックを避けるため nil チェックを徹底

#### 設定パターン
```go
// ハンドラーの初期化
handler := NewHandlerWithStorage(storage)

// MCPツール登録
tool := mcp.NewServerTool(
    "tool_name",
    "Tool description", 
    handler.Method,
    mcp.Input(
        mcp.Property("param1", mcp.Description("Parameter description")),
        mcp.Property("param2", mcp.Description("Optional parameter")),
    ),
)
```

### トラブルシューティング

#### よくある問題

1. **Firestoreアクセスエラー**
   - `GOOGLE_APPLICATION_CREDENTIALS` の設定確認
   - Service Account の権限確認
   - プロジェクトIDの確認

2. **MCPツールが認識されない**
   - Claude Desktop の設定ファイル確認
   - バイナリパスの確認
   - `make install` の実行確認

3. **テスト失敗**
   - `go mod tidy` でモジュール更新
   - `make fmt` でコード整形
   - 環境変数の設定確認

#### デバッグ方法
```bash
# MCP サーバーのログ確認
export MCP_DEBUG=1
memoya

# テスト実行時の詳細ログ
go test -v -race ./...

# 静的解析での詳細確認
go vet ./...
```

## Cloud Run対応アーキテクチャ

### OpenAPI仕様管理

#### 仕様書ベース開発
- **api/openapi.yaml**: 完全なAPI仕様定義
- **自動コード生成**: oapi-codegenによるserver/client生成
- **型安全**: Go構造体自動生成
- **ドキュメント化**: SwaggerUI対応

#### コード生成設定
```yaml
# api/server-config.yaml
package: generated
output: internal/generated/server.go
generate:
  chi-server: true
  models: true
  embedded-spec: true
```

### HTTP Server実装

#### Chi Router + 生成コード統合
```go
type Server struct {
    memoHandler   *handlers.MemoHandler
    todoHandler   *handlers.TodoHandler
    searchHandler *handlers.SearchHandler
    tagHandler    *handlers.TagHandler
}

func (s *Server) CreateMemo(w http.ResponseWriter, r *http.Request) {
    var req generated.MemoCreateRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 既存MCPハンドラーを再利用
    args := handlers.MemoCreateArgs{...}
    params := &mcp.CallToolParamsFor[handlers.MemoCreateArgs]{Arguments: args}
    result, err := s.memoHandler.Create(r.Context(), nil, params)
    
    // JSONレスポンス返却
    w.Write([]byte(result.Content[0].(*mcp.TextContent).Text))
}
```

#### 既存ハンドラーとの統合
- MCPハンドラーをHTTP層でラップ
- ビジネスロジック再利用
- 統一されたエラーハンドリング
- CORS・認証対応

### MCP Client HTTP Transport

#### HTTPブリッジ実装
```go
type MCPBridge struct {
    httpClient *HTTPClient
}

func (b *MCPBridge) MemoCreate(ctx context.Context, ss *mcp.ServerSession, 
    params *mcp.CallToolParamsFor[handlers.MemoCreateArgs]) (*mcp.CallToolResultFor[handlers.MemoCreateResult], error) {
    
    // Cloud Run APIを呼び出し
    respData, err := b.httpClient.CallTool(ctx, "memo_create", params.Arguments)
    
    // MCPレスポンス形式に変換
    return &mcp.CallToolResultFor[handlers.MemoCreateResult]{
        Content: []mcp.Content{&mcp.TextContent{Text: string(respData)}},
    }, nil
}
```

#### 接続性・エラーハンドリング
- サーバーping機能
- 構造化エラーレスポンス
- タイムアウト・リトライ
- 認証トークン管理

### デプロイメント

#### Cloud Run Server
```dockerfile
# マルチステージビルド
FROM golang:1.21-alpine AS builder
RUN make generate  # OpenAPI コード生成
RUN go build -o memoya-server ./cmd/memoya-server

FROM alpine:latest
COPY --from=builder /app/memoya-server .
EXPOSE 8080
CMD ["./memoya-server"]
```

#### 環境変数管理
```bash
# Cloud Run Server
PROJECT_ID=your-firebase-project-id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# MCP Client
MEMOYA_CLOUD_RUN_URL=https://memoya-server-xxx.a.run.app
MEMOYA_AUTH_TOKEN=your-jwt-token  # オプション
```

### 開発ワークフロー

#### コード生成・ビルド
```bash
# 1. 依存関係取得
go mod tidy

# 2. OpenAPIからコード生成
make generate

# 3. テスト実行
make test

# 4. ローカル開発
make run-server  # 別ターミナル
make run-client  # MCPクライアント

# 5. Cloud Runデプロイ
make docker-build
gcloud run deploy memoya-server --image memoya-server
```

#### API仕様変更フロー
1. `api/openapi.yaml`を更新
2. `make generate`でコード再生成
3. サーバー実装を調整
4. テスト実行・動作確認
5. デプロイ

### Cloud Run対応の利点

#### スケーラビリティ
- 自動スケーリング（0〜N台）
- コンカレンシー制御
- リージョン分散対応

#### 可用性・保守性
- 24/7稼働保証
- ゼロダウンタイムデプロイ
- サーバー管理不要

#### セキュリティ
- Google Cloudセキュリティ
- HTTPS強制
- IAM統合認証

#### コスト効率
- 使用量ベース課金
- アイドル時0コスト
- 従量制スケーリング

## マイグレーション戦略

### データ移行の概要

既存のレガシーデータ（ユーザー分離前のデータ）を新しいユーザー分離構造に移行するための包括的なマイグレーション機能を実装済み。

#### 移行対象データ構造

**レガシー構造:**
```
/memos/{memoId}     # 直接配置
/todos/{todoId}     # 直接配置
```

**新構造:**
```
/users/{userId}/memos/{memoId}  # ユーザー分離
/users/{userId}/todos/{todoId}  # ユーザー分離
```

### マイグレーション機能

#### 1. MigrationService実装
- **CheckMigrationStatus**: レガシーデータの存在確認
- **PerformMigration**: 自動データ移行実行
- **CleanupLegacyCollections**: 移行後のクリーンアップ

#### 2. MigrationHandler（MCPツール）
- **migration_status**: 移行状況の確認
- **perform_migration**: 移行の実行
- **cleanup_migration**: レガシーデータの削除

#### 3. 移行プロセス

1. **事前チェック**
   - レガシーデータ数の確認
   - 既存移行状況の確認

2. **デフォルトユーザー作成**
   - 移行用ユーザーの作成または取得
   - Google IDによる既存ユーザー検索

3. **バッチ移行**
   - 500件ずつバッチ処理
   - 原子的操作（作成→削除）
   - 進捗追跡とエラーハンドリング

4. **整合性確認**
   - 移行完了の検証
   - データ整合性チェック

#### 4. 安全機能

- **明示的確認**: 移行・削除操作は`confirm: true`が必須
- **ロールバック不可**: 移行は一方向のみ（安全確保）
- **エラー追跡**: 詳細なエラーログとステータス保存
- **進捗表示**: リアルタイム移行進捗の確認

#### 5. 使用例

```bash
# 1. 移行状況確認
mcp call migration_status

# 2. 移行実行
mcp call perform_migration '{"default_user_google_id":"user@example.com","confirm":true}'

# 3. クリーンアップ（オプション）
mcp call cleanup_migration '{"confirm":true}'
```

### 今後の拡張予定

#### 近期実装
- **Web UI**: SPA + REST API
- **CI/CD**: GitHub Actions
- **監視機能**: Cloud Monitoring統合

#### 中期実装
- **リアルタイム**: WebSocket
- **分析機能**: 使用統計
- **バックアップ**: 自動データバックアップ

#### 長期実装
- **プラグイン**: 拡張アーキテクチャ
- **AI統合**: 自動提案・分類
- **エンタープライズ**: SSO・監査ログ

### 参考資料

#### 外部依存関係
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - MCPプロトコル実装
- [Firebase Admin SDK](https://firebase.google.com/docs/admin/setup) - Firestore操作
- [oapi-codegen](https://github.com/deepmap/oapi-codegen) - OpenAPIコード生成
- [Chi Router](https://github.com/go-chi/chi) - HTTP router
- [Google UUID](https://github.com/google/uuid) - 一意ID生成

#### 関連ドキュメント
- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Claude Desktop MCP Guide](https://claude.ai/docs/mcp)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Firestore Data Model](https://firebase.google.com/docs/firestore/data-model)