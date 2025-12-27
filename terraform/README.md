# Terraform - SUUMO Hunter

## ディレクトリ構成

```
terraform/
├── modules/
│   └── suumo-hunter/      # 共通モジュール
├── _example/              # 設定例（コピーして使用）
│   ├── main.tf
│   └── terraform.tfvars.example
├── shibuya/               # 例: 渋谷エリア用（自分で作成）
│   ├── main.tf
│   └── terraform.tfvars
└── README.md
```

## セットアップ手順

### 1. ビルド

```bash
# プロジェクトルートで
make build
```

### 2. 環境ディレクトリ作成

```bash
# _exampleをコピー
cp -r terraform/_example terraform/shibuya

# terraform.tfvarsを作成
cd terraform/shibuya
cp terraform.tfvars.example terraform.tfvars

# 設定を編集
vim terraform.tfvars
```

### 3. デプロイ

```bash
cd terraform/shibuya
terraform init
terraform plan
terraform apply
```

## 複数インスタンスの運用

### 1つ目のインスタンス（例: 渋谷）

```hcl
# terraform/shibuya/terraform.tfvars
instance_name   = "shibuya"
create_iam_role = true  # IAMロールを作成
```

### 2つ目以降のインスタンス（例: 新宿）

```hcl
# terraform/shinjuku/terraform.tfvars
instance_name   = "shinjuku"
create_iam_role = false  # 既存のIAMロールを使用
```

## リソース命名規則

| リソース | 命名パターン |
|---------|-------------|
| Lambda | `suumo-hunter-{instance}` |
| S3 | `suumo-hunter-{instance}-properties-{account_id}` |
| EventBridge | `suumo-hunter-{instance}-schedule` |
| CloudWatch Logs | `/aws/lambda/suumo-hunter-{instance}` |
| IAMロール | `suumo-hunter-lambda-role`（共通） |

## 環境変数での実行（CI/CD向け）

```bash
export TF_VAR_instance_name="shibuya"
export TF_VAR_suumo_search_url="https://suumo.jp/..."
export TF_VAR_discord_webhook_url="https://discord.com/api/webhooks/..."

terraform apply
```
