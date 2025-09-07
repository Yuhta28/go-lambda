# Aurora PostgreSQL接続監視システム

Amazon Aurora PostgreSQLへのユーザー接続を監視し、CloudWatch Logsの内容をSlackに通知するSAMアプリケーションです。

## 機能

- Aurora PostgreSQLの接続ログをCloudWatch Logsから監視
- 新しいデータベース接続をリアルタイムで検出
- 接続情報（ユーザー名、データベース名、クライアントIP、接続時刻）をSlackに通知
- Go言語で実装されたLambda関数（Amazon Linux 2023ランタイム）

## プロジェクト構造

```bash
.
├── Makefile                    <-- ビルドとデプロイの自動化
├── README.md                   <-- このファイル
├── db-connect-audit           <-- Lambda関数のソースコード
│   ├── main.go                 <-- メインのLambda関数
│   ├── main_test.go            <-- ユニットテスト
│   ├── go.mod                  <-- Go モジュール定義
│   └── go.sum                  <-- 依存関係のチェックサム
├── template.yaml               <-- SAMテンプレート
└── samconfig.toml              <-- SAM設定ファイル
```

## 前提条件

* AWS CLIが管理者権限で設定済み
* [Docker](https://www.docker.com/community-edition)がインストール済み
* [Go言語](https://golang.org) 1.21以上
* [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)がインストール済み
* Aurora PostgreSQLクラスターが作成済み
* CloudWatch Logsが有効化されたAurora PostgreSQL
* Slack Webhook URLの取得済み

## セットアップ手順

### 1. 依存関係のインストールとビルド

SAMの組み込み機能を使用して、依存関係を自動的にダウンロードし、ビルドターゲットをパッケージ化します。

```shell
make build
```

### 2. テストの実行

```shell
make test
```

### 3. デプロイ

初回デプロイ時は以下のコマンドを実行してください：

```shell
make deploy
```

または、パラメータを指定してデプロイ：

```shell
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
AURORA_CLUSTER_ID="your-aurora-cluster" \
LOG_GROUP_NAME="/aws/rds/cluster/your-cluster/postgresql" \
make deploy-prod
```

## 設定パラメータ

デプロイ時に以下のパラメータを設定してください：

| パラメータ名 | 説明 | 例 |
|-------------|------|-----|
| `SlackWebhookUrl` | Slack通知用のWebhook URL | `https://hooks.slack.com/services/...` |
| `AuroraClusterIdentifier` | Aurora PostgreSQLクラスターの識別子 | `aurora-postgresql-cluster` |
| `CloudWatchLogGroupName` | CloudWatch Logsのログループ名 | `/aws/rds/cluster/aurora-postgresql-cluster/postgresql` |

## Aurora PostgreSQLの設定

Aurora PostgreSQLで接続ログを有効にするには、以下の設定が必要です：

1. **パラメータグループの設定**
   ```
   log_connections = 1
   log_disconnections = 1
   log_statement = 'all'  # オプション
   ```

2. **CloudWatch Logsの有効化**
   - RDSコンソールでクラスターを選択
   - 「設定」タブ → 「ログのエクスポート」
   - 「PostgreSQL log」を有効化

## Slackの設定

1. **Slack Appの作成**
   - [Slack API](https://api.slack.com/apps)にアクセス
   - 「Create New App」をクリック
   - 「From scratch」を選択

2. **Incoming Webhookの有効化**
   - 「Incoming Webhooks」を選択
   - 「Activate Incoming Webhooks」をオンに設定
   - 「Add New Webhook to Workspace」をクリック
   - 通知先チャンネルを選択

3. **Webhook URLの取得**
   - 生成されたWebhook URLをコピー
   - デプロイ時のパラメータとして使用

## 通知例

Slackに送信される通知の例：

```
🔗 Aurora PostgreSQLへの新しい接続が検出されました

データベース接続情報
ユーザー名: myuser
データベース名: mydb
クライアントIP: 192.168.1.100
クラスターID: aurora-postgresql-cluster
接続時刻: 2024-01-01 12:00:00 JST
```

## トラブルシューティング

### よくある問題

1. **CloudWatch Logsイベントが発生しない**
   - Aurora PostgreSQLのログ設定を確認
   - CloudWatch Logsのログストリームが作成されているか確認
   - パラメータグループの設定を確認

2. **Slack通知が送信されない**
   - Webhook URLが正しく設定されているか確認
   - Lambda関数のログを確認（CloudWatch Logs）
   - ネットワーク接続を確認

3. **ログの解析に失敗する**
   - PostgreSQLのログフォーマットを確認
   - 正規表現パターンがログ形式と一致しているか確認

### ログの確認

Lambda関数のログは以下で確認できます：
```shell
sam logs -n DbConnectAuditFunction --stack-name your-stack-name --tail
```
## セキュリティ考慮事項

- Slack Webhook URLは機密情報として扱い、環境変数やAWS Systems Manager Parameter Storeに保存
- Lambda関数には最小限の権限のみを付与
- CloudWatch Logsの保持期間を適切に設定
- VPC内のAuroraクラスターの場合、Lambda関数のVPC設定を検討

## カスタマイズ

### ログパターンの変更
`main.go`の`parseConnectionLog`関数で正規表現パターンを変更できます。

### 通知内容の変更
`sendSlackNotification`関数でSlackメッセージの内容をカスタマイズできます。

### フィルタリング
特定のユーザーやIPアドレスをフィルタリングする場合は、`handler`関数に条件を追加してください。

## 参考資料

- [AWS SAM Developer Guide](https://docs.aws.amazon.com/serverless-application-model/)
- [Amazon Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/)
- [Slack API Documentation](https://api.slack.com/)
- [Go AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html)
