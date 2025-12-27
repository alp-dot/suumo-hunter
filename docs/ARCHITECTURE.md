# アーキテクチャ

## システム構成図

```
┌─────────────────────────────────────────────────────────────────┐
│                      AWS Infrastructure                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐     │
│  │ EventBridge │ ───▶ │   Lambda    │ ───▶ │     S3      │     │
│  │  (Cron)     │      │  (Go ARM64) │      │  (CSV保存)  │     │
│  └─────────────┘      └──────┬──────┘      └─────────────┘     │
│                              │                                  │
│                              ├─── 環境変数: DISCORD_WEBHOOK_URL │
│                              │                                  │
│                              ▼                                  │
│            ┌─────────────┐       ┌─────────────┐               │
│            │   Discord   │       │ CloudWatch  │               │
│            │   Webhook   │       │    Logs     │               │
│            └─────────────┘       └─────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

## コンポーネント

### Lambda関数 (Go ARM64)
- ランタイム: `provided.al2023`
- アーキテクチャ: ARM64 (Graviton2)
- メモリ: 256MB
- タイムアウト: 5分

### S3バケット
- 物件データをCSV形式で保存
- バージョニング有効
- 30日後に古いバージョンを自動削除

### EventBridge
- cron式でLambdaを定期実行
- デフォルト: 09:15, 15:15, 18:15, 22:15 (JST)

### Discord Webhook
- 新着物件の通知先
- 2000文字制限のため、長いメッセージは分割送信

## 処理フロー

```
1. EventBridge → Lambda起動
         │
         ▼
2. S3から前回データ(CSV)ダウンロード
         │
         ▼
3. SUUMOスクレイピング（最大30ページ）
         │
         ▼
4. 前回データと比較 → 新着物件検出
         │
         ▼
5. 重回帰分析 → 割安度算出
         │
         ▼
6. 新データをS3にアップロード
         │
         ▼
7. Discord Webhookで通知
```

## プロジェクト構成

```
suumo-hunter-go/
├── cmd/
│   └── lambda/
│       └── main.go              # Lambdaエントリポイント
├── internal/
│   ├── config/
│   │   └── config.go            # 設定管理（環境変数）
│   ├── scraper/
│   │   └── suumo.go             # SUUMOスクレイピング
│   ├── storage/
│   │   └── s3.go                # S3操作
│   ├── notifier/
│   │   └── discord.go           # Discord通知
│   ├── analyzer/
│   │   └── regression.go        # 重回帰分析
│   └── models/
│       └── property.go          # 物件データ構造体
├── terraform/
│   ├── modules/
│   │   └── suumo-hunter/        # Terraformモジュール
│   ├── _example/                # 設定例
│   └── {instance}/              # インスタンスごとの設定
├── docs/                        # ドキュメント
├── .github/workflows/           # CI/CD
├── Makefile
└── README.md
```

## リソース命名規則

| リソース | 命名パターン |
|---------|-------------|
| Lambda | `suumo-hunter-{instance}` |
| S3 | `suumo-hunter-{instance}-properties-{account_id}` |
| EventBridge | `suumo-hunter-{instance}-schedule` |
| CloudWatch Logs | `/aws/lambda/suumo-hunter-{instance}` |
| IAMロール | `suumo-hunter-lambda-role`（共通） |

## タグ管理

すべてのリソースに以下のタグが付与される:

| タグ | 説明 |
|-----|------|
| Project | `suumo-hunter` |
| Instance | インスタンス名（例: `nakano`） |
| ManagedBy | `terraform` |
