# SUUMO Hunter - システム仕様書

## 1. 概要

SUUMOの賃貸物件情報を定期的にスクレイピングし、新着物件を検出してLINE通知するシステム。
重回帰分析による割安度判定機能を備え、お得な物件を自動で発見する。

## 2. システムアーキテクチャ

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
│                              ├─── 環境変数: LINE_NOTIFY_TOKEN   │
│                              │                                  │
│                              ▼                                  │
│            ┌─────────────┐       ┌─────────────┐               │
│            │ LINE Notify │       │ CloudWatch  │               │
│            │    API      │       │    Logs     │               │
│            └─────────────┘       └─────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

## 3. 処理フロー

1. EventBridgeが定期的にLambdaを起動
2. S3から前回取得した物件データ（CSV）をダウンロード
3. SUUMOの検索結果ページをスクレイピング（最大30ページ）
4. 前回データと比較して差分（新着物件）を検出
5. 取得データに対して重回帰分析を実行し、割安度を算出
6. 新しい物件データをS3にアップロード（CSV形式）
7. LINE Notifyで新着物件を通知（割安度付き）

## 4. 機能要件

### 4.1 スクレイピング機能

SUUMOの賃貸物件一覧ページから以下の情報を取得する。

| フィールド | 説明 | データ型 |
|-----------|------|----------|
| name | 物件名 | string |
| address | 住所 | string |
| age | 築年数 | string → int (パース) |
| floor | 階数 | string → int (パース) |
| rent | 家賃 | string → float64 (円) |
| management_fee | 管理費 | string → float64 (円) |
| deposit | 敷金 | string |
| key_money | 礼金 | string |
| layout | 間取り | string |
| area | 専有面積 | string → float64 (m²) |
| walk_minutes | 駅徒歩分数 | int |
| url | 物件詳細URL | string |
| id | 物件ID | string |

※ SUUMOから取得した「万円」表記は10,000を乗じて円に変換する。例: 「7.9万円」→ 79,000円

#### 駅徒歩分数の取得ルール
- 複数路線が表示されている場合は、最初に表示されている駅（最寄り駅）の徒歩分数を採用
- 例: 「新井薬師前駅 歩8分 / 沼袋駅 歩10分」→ 8分を採用

#### 物件IDの形式
- SUUMOの物件詳細URLに含まれる `jnc_XXXXXXXXXXXX` 形式のIDを使用
- 例: `/chintai/jnc_000102396492/` → `jnc_000102396492`

### 4.2 重回帰分析機能

#### 目的変数
- 総賃料（rent + management_fee）

#### 説明変数
- 専有面積（area）
- 築年数（age）
- 階数（floor）
- 駅徒歩分数（walk_minutes）

#### お得度の算出

```
お得度（円） = 予測総賃料 - 実際総賃料
```

- 正の値: 相場より安い（お得）
- 負の値: 相場より高い（割高）
- 判定基準:
  - +10,000円以上: お買い得
  - -10,000円〜+10,000円: 標準
  - -10,000円未満: 割高

#### 分析の前提条件

- 最低サンプル数: 10件以上
- サンプル不足時: 回帰分析をスキップし、お得度を算出しない（通知時は「分析中」と表示）

### 4.3 通知機能

LINE Notify APIを使用して以下の形式で通知する。

```
🏠 新着物件のお知らせ

▫️マンション名A
📍 東京都渋谷区...
💰 8.5万円（管理費込）
💴 相場より 12,800円/月 お得！
🔗 https://suumo.jp/...

▫️マンション名B
📍 東京都新宿区...
💰 9.0万円（管理費込）
💴 相場より 2,100円/月 お得
🔗 https://suumo.jp/...

▫️マンション名C
📍 東京都目黒区...
💰 10.5万円（管理費込）
💴 相場より 8,500円/月 高い
🔗 https://suumo.jp/...
```

#### 通知の制限

- 1回の通知上限: 10件（超過分は「他N件の新着あり」と要約）
- メッセージ長制限: 1000文字を超える場合は分割送信

### 4.4 データ永続化

- 形式: CSV
- 保存先: AWS S3
- 重複排除キー: id（物件ID）

## 5. 非機能要件

### 5.1 パフォーマンス

- スクレイピング: リトライ3回、10秒間隔、指数バックオフ
- Lambda実行時間: 最大15分
- メモリ: 256MB（調整可能）

### 5.2 エラーハンドリング

- スクレイピングエラー: 3回リトライ後も失敗した場合はCloudWatch Logsにエラー記録し終了
- S3書き込み失敗: 通知は行わず、次回実行時に再試行
- LINE通知失敗: CloudWatch Logsに記録し、処理は継続（データは保存）

### 5.3 セキュリティ

- LINE Notify Token: 環境変数
- IAMロール: 最小権限の原則

### 5.4 可用性

- 定期実行: EventBridge Scheduler
- エラー時: CloudWatch Logsに記録

### 5.5 品質基準

#### コードスタイル
- golangci-lint を使用
- Google Go Style Guide に準拠
- CI で自動チェック

#### テスト
- 主要ロジックに対するユニットテストを実装
- テスト対象:
  - `internal/scraper` - HTMLパース、データ抽出
  - `internal/analyzer` - 重回帰分析、お得度計算
  - `internal/models` - データ変換、バリデーション
  - `internal/storage` - S3操作（モックを使用）
  - `internal/notifier` - LINE通知（モックを使用）
- カバレッジ目標: 主要ロジック80%以上

#### CI
- GitHub Actions でパイプライン構築
- プルリクエスト時:
  - `golangci-lint run`
  - `go test ./...`

#### デプロイ
- Terraformはローカル実行
- `make deploy` でビルド → terraform apply

## 6. プロジェクト構成

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
│   │   └── line.go              # LINE通知
│   ├── analyzer/
│   │   └── regression.go        # 重回帰分析・割安度判定
│   └── models/
│       └── property.go          # 物件データ構造体
├── terraform/
│   ├── main.tf                  # プロバイダー設定
│   ├── variables.tf             # 変数定義
│   ├── outputs.tf               # 出力定義
│   ├── lambda.tf                # Lambda関数
│   ├── s3.tf                    # S3バケット
│   ├── eventbridge.tf           # 定期実行スケジュール
│   └── iam.tf                   # IAMロール・ポリシー
├── .github/
│   └── workflows/
│       └── ci.yml               # CI/CDパイプライン
├── .golangci.yml                # golangci-lint設定
├── Makefile                     # ビルド・デプロイコマンド
├── go.mod
├── go.sum
├── SPEC.md                      # 本仕様書
└── README.md
```

## 7. 設定・環境変数

### 7.1 環境変数（Lambda設定）

| 変数名 | 説明 | 必須 |
|--------|------|------|
| BUCKET_NAME | S3バケット名 | ✓ |
| BUCKET_KEY | CSVファイルのキー | - (default: properties.csv) |
| MAX_PAGE | スクレイピング最大ページ数 | - (default: 30) |
| SUUMO_SEARCH_URL | SUUMO検索URL | ✓ |
| LINE_NOTIFY_TOKEN | LINE Notify APIトークン | ✓ |

## 8. 依存ライブラリ

### 8.1 Go パッケージ

| パッケージ | 用途 |
|-----------|------|
| github.com/aws/aws-lambda-go | Lambdaランタイム |
| github.com/aws/aws-sdk-go-v2 | AWS SDK (S3) |
| github.com/PuerkitoBio/goquery | HTMLスクレイピング |
| github.com/avast/retry-go | リトライ処理 |
| github.com/caarlos0/env/v9 | 環境変数パース |
| gonum.org/v1/gonum/stat | 重回帰分析 |

### 8.2 Terraform プロバイダー

| プロバイダー | バージョン |
|-------------|-----------|
| hashicorp/aws | ~> 5.0 |

## 9. 開発フェーズ

### Phase 1: 基本機能のGo移植

- [ ] プロジェクト構成・go.mod初期化
- [ ] 物件データ構造体定義
- [ ] スクレイピング機能 (goquery)
- [ ] S3操作 (aws-sdk-go-v2)
- [ ] LINE通知機能
- [ ] Lambdaハンドラ実装

### Phase 2: インフラ整備

- [ ] Terraform基盤構築
- [ ] S3バケット作成
- [ ] Lambda関数デプロイ
- [ ] IAMロール・ポリシー設定
- [ ] EventBridgeスケジュール設定

### Phase 3: 分析機能追加

- [ ] データクレンジング（築年数・面積の数値化）
- [ ] 重回帰分析実装
- [ ] 割安度スコア算出
- [ ] 通知フォーマット改善

## 10. 補足

### 10.1 SUUMO検索URL例

```
https://suumo.jp/jj/chintai/ichiran/FR301FC001/?ar=030&bs=040&pc=20&smk=&po1=25&po2=99&shkr1=03&shkr2=03&shkr3=03&shkr4=03&rn=0350&ek=035001440&ra=013&cb=0.0&ct=11.0&co=1&et=9999999&mb=0&mt=9999999&cn=20&fw2=&page=
```

#### 主要パラメータ

| パラメータ | 説明 | 例 |
|-----------|------|-----|
| ar | エリアコード | 030 (関東) |
| bs | 物件種別 | 040 (賃貸) |
| pc | 1ページあたり表示件数 | 20 |
| cb | 賃料下限（万円） | 0.0 |
| ct | 賃料上限（万円） | 11.0 |
| cn | 築年数上限（年） | 20 |
| co | 管理費込みフラグ | 1 (込み) |
| rn | 路線コード | 0350 (西武新宿線) |
| ek | 駅コード | 035001440 (新井薬師前) |
| page | ページ番号 | 1〜 |

### 10.2 LINE Notify API

- エンドポイント: `https://notify-api.line.me/api/notify`
- 認証: Bearer Token
- メソッド: POST
- Content-Type: application/x-www-form-urlencoded
