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
- Go 1.24.4以上
- Firebase プロジェクト（Firestore使用）

### インストール

1. リポジトリをクローン
```bash
git clone <repository-url>
cd memoya
```

2. 依存関係をインストール
```bash
go mod download
```

3. 環境変数を設定
```bash
export FIREBASE_PROJECT_ID=your-firebase-project-id
```

4. ビルド
```bash
go build ./cmd/memoya
```

## 使用方法

### MCPサーバーとして起動
```bash
./memoya
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

AIクライアントから以下のようにツールを呼び出します：

```json
{
  "tool": "todo_create",
  "arguments": {
    "title": "プロジェクトの企画書を作成",
    "description": "来週のミーティング用の企画書を準備する",
    "status": "todo",
    "priority": "high",
    "tags": ["work", "urgent"]
  }
}
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

## 設定

### 環境変数

- `FIREBASE_PROJECT_ID`: Firebase プロジェクトID（必須）

### Firebase設定

1. [Firebase Console](https://console.firebase.google.com/) でプロジェクトを作成
2. Firestoreを有効化
3. サービスアカウントキーを生成して適切に設定

## ライセンス

MIT License

## 貢献

プルリクエストや Issue の作成を歓迎します。

## TODO

- [ ] インメモリストレージの実装（開発用）
- [ ] より詳細な検索機能
- [ ] バルク操作のサポート
- [ ] テストケースの追加
- [ ] パフォーマンスの最適化