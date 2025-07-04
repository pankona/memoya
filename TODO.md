# TODO / 改善案

## アーキテクチャ改善

### Cloud Functions化による利便性向上
- **現在の課題**: memoyaを動かす環境で秘密鍵とproject idを用意する必要があり不便
- **改善案**: Cloud Functionsなどのサーバーレス環境にAPIを配置し、MCPクライアントからHTTP経由で呼び出す構成に変更
- **メリット**:
  - クライアント側で秘密鍵管理が不要
  - 複数環境での利用が簡単
  - セキュリティ向上（秘密鍵の分散を避けられる）
  - デプロイ・運用の簡素化

### 実装時の検討事項
- Cloud Functions + Firestore の構成
- 認証方式の検討（API Key、OAuth など）
- MCP over HTTP の実装
- エラーハンドリングとレスポンス形式の統一