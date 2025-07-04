# Memoya Cloud Run Deployment Guide

このドキュメントでは、MemoやのGoogle Cloud Runへのデプロイメント手順を説明します。

## 前提条件

### 必要なツール
- [Google Cloud CLI (gcloud)](https://cloud.google.com/sdk/docs/install)
- [Docker](https://docs.docker.com/get-docker/)
- Git

### 必要な権限
- Cloud Run Admin
- Service Account Admin
- Secret Manager Admin
- Cloud Build Editor
- Storage Admin (Container Registry用)

## クイックスタート

### 1. 環境準備

```bash
# Google Cloud にログイン
gcloud auth login

# プロジェクトIDを設定
export PROJECT_ID=your-project-id
gcloud config set project $PROJECT_ID

# 必要なAPIを有効化
gcloud services enable cloudbuild.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable firestore.googleapis.com
```

### 2. OAuth認証設定

Google Cloud Console > APIs & Services > Credentialsで：

1. OAuth 2.0 Client IDを作成
2. Application typeを「Desktop application」に設定
3. Client IDとClient Secretをメモ

### 3. Secret Manager設定

```bash
# Secret Manager設定スクリプトを実行
./scripts/setup-secrets.sh

# または手動で設定
gcloud secrets create oauth-client-id --replication-policy="automatic"
gcloud secrets create oauth-client-secret --replication-policy="automatic"

# 値を設定
echo "your-oauth-client-id" | gcloud secrets versions add oauth-client-id --data-file=-
echo "your-oauth-client-secret" | gcloud secrets versions add oauth-client-secret --data-file=-
```

### 4. デプロイ実行

```bash
# 自動デプロイ（推奨）
./scripts/deploy.sh

# または手動デプロイ
./scripts/deploy.sh --manual
```

## 詳細設定

### 環境変数設定

Cloud Runで設定する環境変数：

| 変数名 | 必須 | 説明 | 例 |
|--------|------|------|-----|
| `PROJECT_ID` | ✅ | Firebase/GCPプロジェクトID | `memoya-prod` |
| `CORS_ALLOWED_ORIGINS` | ❌ | CORS許可オリジン（カンマ区切り） | `https://memoya.example.com,https://app.memoya.com` |
| `PORT` | ❌ | サーバーポート（通常8080） | `8080` |

### Service Account権限

Service Account `memoya-server@PROJECT_ID.iam.gserviceaccount.com` に必要な権限：

- `roles/firestore.user` - Firestoreへのアクセス
- `roles/secretmanager.secretAccessor` - Secret Managerからの読み取り

### Firestore設定

1. Firebase Console > Firestore Database > Create Database
2. Location: `asia-northeast1` (推奨)
3. Security rules を設定（認証済みユーザーのみアクセス可能）

## デプロイメント方法

### 方法1: Cloud Build（推奨）

```bash
# cloudbuild.yamlを使用した自動ビルド・デプロイ
gcloud builds submit --config cloudbuild.yaml .
```

**特徴：**
- ✅ 完全自動化
- ✅ イメージキャッシュによる高速ビルド
- ✅ 複数環境対応
- ✅ ロールバック機能

### 方法2: 手動デプロイ

```bash
# Dockerイメージをローカルでビルド
docker build -t gcr.io/$PROJECT_ID/memoya-server .
docker push gcr.io/$PROJECT_ID/memoya-server

# Cloud Runにデプロイ
gcloud run deploy memoya-server \
  --image gcr.io/$PROJECT_ID/memoya-server \
  --region asia-northeast1 \
  --platform managed \
  --allow-unauthenticated
```

## Cloud Run設定詳細

### リソース設定

| 設定項目 | 推奨値 | 説明 |
|----------|--------|------|
| CPU | 1 | vCPU数 |
| Memory | 512Mi | メモリ上限 |
| Min instances | 0 | 最小インスタンス数（コスト最適化） |
| Max instances | 10 | 最大インスタンス数 |
| Concurrency | 80 | 同時リクエスト数 |
| Timeout | 300s | リクエストタイムアウト |

### ネットワーク設定

- **Ingress**: All traffic
- **Authentication**: Allow unauthenticated invocations
- **Port**: 8080

## モニタリング・ログ

### Cloud Logging

ログは自動的にCloud Loggingに送信されます：

```bash
# ログ確認
gcloud logs read "resource.type=cloud_run_revision AND resource.labels.service_name=memoya-server" --limit 50

# リアルタイムログ
gcloud logs tail "resource.type=cloud_run_revision AND resource.labels.service_name=memoya-server"
```

### ヘルスチェック

- **エンドポイント**: `https://your-service-url/health`
- **期待レスポンス**: `{"status":"ok","timestamp":"..."}`

### メトリクス

Cloud Monitoringで監視可能：
- Request count
- Request latency
- Instance count
- CPU usage
- Memory usage

## トラブルシューティング

### よくある問題

#### 1. デプロイに失敗する

```bash
# ビルドログ確認
gcloud builds list --limit=5

# 詳細ログ確認
gcloud builds log BUILD_ID
```

**解決方法：**
- Dockerfile構文確認
- 依存関係の問題確認
- Cloud Build権限確認

#### 2. 認証エラー

```
Error: failed to get OAuth credentials
```

**解決方法：**
- Secret Manager設定確認
- Service Account権限確認
- OAuth設定確認

#### 3. Firestoreアクセスエラー

```
Error: failed to initialize Firestore
```

**解決方法：**
- PROJECT_ID環境変数確認
- Service Account権限確認
- Firestore API有効化確認

#### 4. CORS エラー

```
Access to fetch at 'https://...' from origin '...' has been blocked by CORS policy
```

**解決方法：**
- `CORS_ALLOWED_ORIGINS`環境変数設定
- オリジンURL確認

### ログ分析

```bash
# エラーログのみ表示
gcloud logs read "resource.type=cloud_run_revision AND resource.labels.service_name=memoya-server AND severity>=ERROR" --limit 20

# 特定時間範囲のログ
gcloud logs read "resource.type=cloud_run_revision AND resource.labels.service_name=memoya-server" \
  --since="2024-01-01T00:00:00Z" --until="2024-01-02T00:00:00Z"
```

## セキュリティ

### ベストプラクティス

1. **Secret Manager使用**
   - 機密情報はSecret Managerで管理
   - 環境変数に機密情報を直接設定しない

2. **Service Account最小権限**
   - 必要最小限の権限のみ付与
   - 定期的な権限見直し

3. **CORS設定**
   - プロダクション環境では特定ドメインのみ許可
   - ワイルドカード（`*`）は避ける

4. **ネットワークセキュリティ**
   - 必要に応じてCloud Load BalancerでSSL終端
   - Cloud Armorでアクセス制御

### セキュリティ監査

```bash
# Service Account権限確認
gcloud projects get-iam-policy $PROJECT_ID

# Secret Manager権限確認
gcloud secrets get-iam-policy oauth-client-id
gcloud secrets get-iam-policy oauth-client-secret
```

## パフォーマンス最適化

### レスポンス最適化

1. **コールドスタート対策**
   - Min instances > 0 に設定（コスト vs パフォーマンス）
   - 軽量な初期化処理

2. **同時実行数最適化**
   - Concurrency値の調整
   - CPU/Memory使用量に応じて調整

3. **タイムアウト設定**
   - 適切なタイムアウト値設定
   - 長時間処理の分離

## コスト最適化

### 推奨設定

```yaml
# cloudbuild.yaml での設定例
- name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
  args: [
    'run', 'deploy', 'memoya-server',
    '--min-instances', '0',      # コールドスタート許容
    '--max-instances', '5',      # 上限制御
    '--cpu', '1',                # 必要最小限
    '--memory', '512Mi',         # 必要最小限
    '--concurrency', '80'        # 高い同時実行数
  ]
```

### コスト監視

```bash
# 使用量確認
gcloud run services describe memoya-server --region=asia-northeast1

# 課金データ確認（BigQueryが必要）
bq query --use_legacy_sql=false '
SELECT
  service.description,
  SUM(cost) as total_cost
FROM `PROJECT_ID.cloud_billing_export.gcp_billing_export_v1_BILLING_ACCOUNT_ID`
WHERE service.description LIKE "%Cloud Run%"
  AND invoice.month = "202401"
GROUP BY service.description
'
```

## CI/CD統合

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Cloud Run

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Cloud SDK
      uses: google-github-actions/setup-gcloud@v1
      with:
        project_id: ${{ secrets.GCP_PROJECT_ID }}
        service_account_key: ${{ secrets.GCP_SA_KEY }}
        export_default_credentials: true
    
    - name: Build and Deploy
      run: |
        gcloud builds submit --config cloudbuild.yaml .
```

### 必要なSecrets

- `GCP_PROJECT_ID`: プロジェクトID
- `GCP_SA_KEY`: Service AccountのJSONキー

## バックアップ・災害復旧

### Firestoreバックアップ

```bash
# 定期バックアップ設定
gcloud firestore operations list

# Export実行
gcloud firestore export gs://your-backup-bucket/firestore-backup-$(date +%Y%m%d)
```

### イメージバックアップ

Container Registryは自動的にイメージを保持。過去バージョンへのロールバック：

```bash
# 過去のイメージ一覧
gcloud container images list-tags gcr.io/$PROJECT_ID/memoya-server

# 特定バージョンにロールバック
gcloud run deploy memoya-server \
  --image gcr.io/$PROJECT_ID/memoya-server:COMMIT_SHA \
  --region asia-northeast1
```

## 関連ドキュメント

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)
- [Firestore Documentation](https://cloud.google.com/firestore/docs)
- [Cloud Build Documentation](https://cloud.google.com/build/docs)
- [Memoya Architecture Documentation](../CLAUDE.md#cloud-run対応アーキテクチャ)