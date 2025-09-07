# Aurora MySQL 接続監視システム

Amazon Aurora MySQL へのユーザー接続を監視し、CloudWatch Logs の内容を Slack に通知する SAM アプリケーションです。

## 機能

- Aurora MySQL の General Log を監視
- 新しいデータベース接続をリアルタイムで検出
- 接続情報（ユーザー名、データベース名、クライアント IP、接続時刻）を Slack に通知
- Go 言語で実装された Lambda 関数（Amazon Linux 2023 ランタイム）

## 前提条件

- AWS CLI が管理者権限で設定済み
- [Docker](https://www.docker.com/community-edition)がインストール済み
- [Go 言語](https://golang.org) 1.21 以上
- [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)がインストール済み
- Aurora MySQL クラスターが作成済み
- CloudWatch Logs が有効化された Aurora MySQL
- Slack Webhook URL の取得済み

## Aurora MySQL の設定

Aurora MySQL で接続ログを有効にするには、以下の設定が必要です：

1. **パラメータグループの設定**

   ```
   general_log = 1
   log_output = FILE
   ```

2. **CloudWatch Logs の有効化**
   - RDS コンソールでクラスターを選択
   - 「設定」タブ → 「ログのエクスポート」
   - 「General log」を有効化

## セットアップ手順

### 1. 依存関係のインストールとビルド

```shell
make build
```

### 2. テストの実行

```shell
make test
```

### 3. デプロイ

初回デプロイ時：

```shell
make deploy
```

パラメータを指定してデプロイ：

```shell
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL" \
AURORA_CLUSTER_ID="your-mysql-cluster" \
LOG_GROUP_NAME="/aws/rds/cluster/your-cluster/general" \
make deploy-prod
```

## 設定パラメータ

| パラメータ名              | 説明                            | 例                                              |
| ------------------------- | ------------------------------- | ----------------------------------------------- |
| `SlackWebhookUrl`         | Slack 通知用の Webhook URL      | `https://hooks.slack.com/services/...`          |
| `AuroraClusterIdentifier` | Aurora MySQL クラスターの識別子 | `aurora-mysql-cluster`                          |
| `CloudWatchLogGroupName`  | CloudWatch Logs のログループ名  | `/aws/rds/cluster/aurora-mysql-cluster/general` |

## 対応するログ形式

このシステムは以下の MySQL ログ形式に対応しています：

```
# データベース指定あり
2025-09-07T06:41:11.701820Z	  252 Connect	test28@10.0.139.222 on appdb using TCP/IP

# データベース指定なし
2025-09-07T06:40:32.160890Z	  249 Connect	test28@10.0.139.222 on  using TCP/IP
```

## 通知例

Slack に送信される通知の例：

```
🔗 Aurora MySQLへの新しい接続が検出されました

データベース接続情報
ユーザー名: test28
データベース名: appdb (または「指定なし」)
クライアントIP: 10.0.139.222
クラスターID: aurora-mysql-cluster
接続時刻: 2025-09-07 15:41:11 JST
```
