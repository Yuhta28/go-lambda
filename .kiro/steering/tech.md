# 技術スタック

## ランタイム・言語
- **Go 1.21+** - メインプログラミング言語
- **AWS Lambda** - サーバーレスコンピュート（provided.al2023ランタイム）
- **Amazon Linux 2023** - Lambda実行環境

## AWSサービス
- **AWS SAM** - Infrastructure as CodeのためのServerless Application Model
- **CloudWatch Logs** - ログ監視とイベントトリガー
- **Aurora PostgreSQL** - 監視対象データベース
- **IAM** - アイデンティティ・アクセス管理

## 依存関係
- `github.com/aws/aws-lambda-go` - AWS Lambda Go SDK

## ビルドシステム・コマンド

### 前提条件
- 適切な権限で設定されたAWS CLI
- Dockerのインストール
- Go 1.21+のインストール
- SAM CLIのインストール
- CloudWatch Logsが有効化されたAurora PostgreSQLクラスター
- Slack Webhook URL

### 共通コマンド

```bash
# アプリケーションのビルド
make build

# テストの実行
make test

# SAMテンプレートの検証
make validate

# ガイド付きセットアップでデプロイ（初回）
make deploy

# パラメータ指定での本番デプロイ
SLACK_WEBHOOK_URL="https://hooks.slack.com/..." \
AURORA_CLUSTER_ID="your-cluster" \
LOG_GROUP_NAME="/aws/rds/cluster/your-cluster/postgresql" \
make deploy-prod

# リソースのクリーンアップ
make clean

# Lambdaログの表示
sam logs -n DbConnectAuditFunction --stack-name db-connect-audit --tail
```

## 設定ファイル
- `template.yaml` - AWSリソースを定義するSAMテンプレート
- `samconfig.toml` - SAMデプロイ設定
- `go.mod` - Goモジュール依存関係
- `Makefile` - ビルド自動化