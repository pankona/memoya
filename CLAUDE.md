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

### 全体構成
```
Claude Desktop ↔ MCP Protocol ↔ memoya (Go) ↔ Firestore
```

### ディレクトリ構造
```
memoya/
├── cmd/memoya/           # メインエントリーポイント
├── internal/
│   ├── handlers/         # MCPハンドラー実装
│   │   ├── memo.go      # メモ操作
│   │   ├── todo.go      # TODO操作  
│   │   ├── search.go    # 検索機能
│   │   ├── tag.go       # タグ管理
│   │   ├── *_test.go    # 単体テスト
│   │   └── mock_storage.go # テスト用モック
│   ├── models/          # データモデル定義
│   ├── storage/         # ストレージインターフェース
│   └── config/          # 設定管理
├── Makefile             # ビルド・テストコマンド
└── go.mod               # Go依存関係
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

### 今後の拡張予定

#### 認証機能
- Google OAuth 2.0 Device Flow対応
- マルチユーザー対応
- JWT トークン管理

#### Web UI
- Cloud Run対応
- HTTP transport 対応  
- REST API エンドポイント

#### 機能拡張
- タグ正規化システム
- メモ・TODO間のリンク機能強化
- エクスポート機能（JSON、Markdown）
- 全文検索の高度化

### 参考資料

#### 外部依存関係
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - MCPプロトコル実装
- [Firebase Admin SDK](https://firebase.google.com/docs/admin/setup) - Firestore操作
- [Google UUID](https://github.com/google/uuid) - 一意ID生成

#### 関連ドキュメント
- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Claude Desktop MCP Guide](https://claude.ai/docs/mcp)
- [Firestore Data Model](https://firebase.google.com/docs/firestore/data-model)