# CLAUDE.md

## プロジェクトルール

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