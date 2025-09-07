# プロジェクト構造

## リポジトリ構成

```
.
├── .git/                       # Gitバージョン管理
├── .gitignore                  # Git除外パターン
├── .kiro/                      # Kiro AIアシスタント設定
│   └── steering/               # AI誘導ルールとガイドライン
├── README.md                   # ルートプロジェクトドキュメント
└── db-connect-audit/           # メインアプリケーションディレクトリ
    ├── .aws-sam/               # SAMビルド成果物（生成）
    ├── db-connect-audit/       # Goソースコードディレクトリ
    │   ├── main.go             # Lambda関数実装
    │   ├── main_test.go        # ユニットテスト
    │   ├── go.mod              # Goモジュール定義
    │   └── go.sum              # 依存関係チェックサム
    ├── events/                 # テスト用サンプルイベントデータ
    ├── Makefile                # ビルド・デプロイ自動化
    ├── README.md               # アプリケーション固有ドキュメント
    ├── samconfig.toml          # SAMデプロイ設定
    ├── samconfig.toml.template # SAM設定テンプレート
    └── template.yaml           # SAMインフラストラクチャテンプレート
```

## 主要規約

### ディレクトリ構造
- **ルートレベル**: プロジェクト概要とメインアプリケーションフォルダを含む
- **db-connect-audit/**: 全SAMリソースを含むメインアプリケーション
- **db-connect-audit/db-connect-audit/**: Goソースコード（SAM CodeUriと一致）
- **生成ディレクトリ**: `.aws-sam/`はビルド成果物を含む（コミット対象外）

### ファイル命名
- Goファイルは標準命名を使用: `main.go`, `main_test.go`
- SAMテンプレート: `template.yaml`（標準SAM規約）
- 設定: デプロイ設定用の`samconfig.toml`
- ドキュメント: ルートとアプリケーションレベル両方の`README.md`

### コード構成
- アプリケーションディレクトリごとに単一Lambda関数
- ソースコードと同じ場所に配置されたテスト
- SAMパラメータによる環境固有設定
- SAMテンプレートを使用したInfrastructure as Code

### 設定管理
- 新環境のベースとして`samconfig.toml.template`を使用
- 機密値（Slack webhook）はパラメータとして渡す
- samconfig.tomlでのリージョン固有設定
- SAMテンプレートで定義された環境変数