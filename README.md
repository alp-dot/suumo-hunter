# SUUMO Hunter

SUUMOの賃貸物件を自動でスクレイピングし、新着物件をDiscordに通知するサーバーレスアプリケーション。
重回帰分析による割安度判定機能を備え、お得な物件を自動で発見します。

## 機能

- 🏠 **自動スクレイピング** - SUUMOの検索結果を定期的に取得
- 📊 **割安度分析** - 重回帰分析で相場と比較し、お得な物件を判定
- 🔔 **Discord通知** - 新着物件をリアルタイムでお知らせ
- ☁️ **サーバーレス** - AWS Lambda + EventBridgeで低コスト運用
- 🗂️ **複数エリア対応** - 異なる検索条件で複数のbotを並行運用可能

## クイックスタート

### 前提条件

- [Go](https://golang.org/) 1.22+
- [Terraform](https://www.terraform.io/) 1.0+
- [AWS CLI](https://aws.amazon.com/cli/) v2（設定済み）
- Discord サーバー（通知先）

### 1. リポジトリのクローン

```bash
git clone https://github.com/your-username/suumo-hunter-go.git
cd suumo-hunter-go
```

### 2. Discord Webhook URLの取得

1. Discordでサーバー設定を開く
2. 「連携サービス」→「ウェブフック」
3. 「新しいウェブフック」を作成
4. 通知先のチャンネルを選択
5. **ウェブフックURLをコピー**

### 3. SUUMO検索URLの取得

1. [SUUMO](https://suumo.jp/chintai/) で希望条件を設定して検索
2. 検索結果ページのURLをコピー

例:
```
https://suumo.jp/jj/chintai/ichiran/FR301FC001/?ar=030&bs=040&pc=20&smk=&po1=25&po2=99&shkr1=03&shkr2=03&shkr3=03&shkr4=03&sc=13114&ta=13&cb=0.0&ct=20.0&co=1&et=9999999&mb=0&mt=9999999&cn=9999999&fw2=
```

### 4. ビルド

```bash
make build
```

### 5. Terraform設定

```bash
# 設定ディレクトリを作成（例: 中野エリア）
cp -r terraform/_example terraform/nakano
cd terraform/nakano

# terraform.tfvarsを作成
cp terraform.tfvars.example terraform.tfvars
```

`terraform.tfvars` を編集:

```hcl
instance_name       = "nakano"
suumo_search_url    = "https://suumo.jp/jj/chintai/ichiran/..."  # Step 3でコピーしたURL
discord_webhook_url = "https://discord.com/api/webhooks/..."     # Step 2でコピーしたURL
```

### 6. デプロイ

```bash
terraform init
terraform plan    # 確認
terraform apply   # 実行
```

### 7. 動作確認

```bash
# 手動実行
aws lambda invoke --function-name suumo-hunter-nakano output.json
cat output.json

# ログ確認
aws logs tail /aws/lambda/suumo-hunter-nakano --since 5m
```

## 設定オプション

`terraform.tfvars` で以下の設定が可能:

| 変数 | 説明 | デフォルト |
|------|------|-----------|
| `instance_name` | インスタンス識別子（例: nakano, shibuya） | 必須 |
| `suumo_search_url` | SUUMOの検索URL | 必須 |
| `discord_webhook_url` | Discord Webhook URL | 必須 |
| `aws_region` | AWSリージョン | `ap-northeast-1` |
| `max_page` | スクレイピング最大ページ数 | `30` |
| `schedule_expression` | 実行スケジュール (cron) | `cron(15 0,6,9,13 * * ? *)` |
| `create_iam_role` | IAMロールを作成するか | `true` |

### 実行スケジュール

デフォルトでは以下の時刻（JST）に実行:
- 09:15
- 15:15
- 18:15
- 22:15

変更例:
```hcl
# 毎日9時と21時に実行
schedule_expression = "cron(0 0,12 * * ? *)"  # UTCで指定
```

## 複数エリアの運用

異なる検索条件で複数のbotを並行運用できます。

### 2つ目のインスタンス追加

```bash
# 新しいディレクトリを作成
cp -r terraform/_example terraform/shibuya
cd terraform/shibuya

# terraform.tfvarsを編集
cp terraform.tfvars.example terraform.tfvars
```

```hcl
instance_name       = "shibuya"
suumo_search_url    = "https://suumo.jp/..."  # 渋谷エリアの検索URL
discord_webhook_url = "https://discord.com/..."
create_iam_role     = false  # 2つ目以降はIAMロール共有
```

```bash
terraform init
terraform apply
```

## 開発

### テスト実行

```bash
make test
```

### Lint

```bash
make lint
```

### ビルド (Lambda用)

```bash
make build
```

## アーキテクチャ

```
EventBridge (cron) → Lambda (Go) → S3 (CSV保存)
                          │
                          ├→ Discord Webhook (通知)
                          └→ CloudWatch Logs (ログ)
```

詳細: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## ドキュメント

- [アーキテクチャ](docs/ARCHITECTURE.md) - システム構成詳細
- [重回帰分析](docs/ANALYSIS.md) - 割安度判定の仕組み
- [仕様書](SPEC.md) - 詳細な技術仕様
- [Terraform](terraform/README.md) - インフラ構築詳細

## ライセンス

MIT
