# SUUMO Hunter Go 実装計画

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

本計画は `PLANS.md` に準拠して作成・維持される。

## Purpose / Big Picture

この実装計画は、既存のPython版SUUMOスクレイパーをGoに移植し、AWS Lambda上で動作するシステムを構築することを目的とする。完成後、ユーザーは以下のことが可能になる：

1. SUUMOの賃貸物件情報が定期的にスクレイピングされ、新着物件がDiscord通知で受け取れる
2. 重回帰分析による割安度判定により、相場より安い物件を自動で発見できる
3. AWS EventBridgeによる定期実行で、手動操作なしに継続的に物件情報を監視できる

動作確認方法：Lambdaをデプロイ後、手動実行またはスケジュール実行により、Discordに新着物件通知が届くことを確認する。ローカルでは `go run cmd/lambda/main.go` でテスト実行できる（環境変数設定が必要）。

## Progress

- [x] (2025-12-27) Milestone 1: プロジェクト基盤構築
  - [x] Go module初期化（go.mod）
  - [x] ディレクトリ構造作成
  - [x] 依存ライブラリ追加
  - [x] Makefile作成
  - [x] golangci-lint設定（.golangci.yml）

- [x] (2025-12-27) Milestone 2: データモデル実装
  - [x] Property構造体定義（internal/models/property.go）
  - [x] CSV読み書き機能（internal/models/csv.go）
  - [x] データ変換ユーティリティ（万円→円変換など）
  - [x] ユニットテスト作成（カバレッジ90.2%）

- [x] (2025-12-27) Milestone 3: スクレイピング機能実装
  - [x] SUUMOスクレイパー実装（internal/scraper/suumo.go）
  - [x] HTMLパース処理（goquery使用）
  - [x] ページネーション対応
  - [x] リトライ処理（指数バックオフ、retry-go使用）
  - [x] ユニットテスト作成（カバレッジ94.6%）

- [x] (2025-12-27) Milestone 4: ストレージ機能実装
  - [x] S3クライアント実装（internal/storage/s3.go）
  - [x] CSVアップロード/ダウンロード
  - [x] ユニットテスト作成（モック使用、カバレッジ86.2%）

- [x] (2025-12-27) Milestone 5: 通知機能実装
  - [x] Discord Webhookクライアント実装（internal/notifier/discord.go）
  - [x] メッセージフォーマット（SPEC.md準拠）
  - [x] 文字数制限・分割送信（1000文字、10件上限）
  - [x] ユニットテスト作成（モック使用、カバレッジ92.6%）

- [x] (2025-12-27) Milestone 6: 分析機能実装
  - [x] 重回帰分析実装（internal/analyzer/regression.go）
  - [x] 割安度算出ロジック（gonum/mat使用）
  - [x] ユニットテスト作成（カバレッジ88.3%）

- [x] (2025-12-27) Milestone 7: Lambda統合
  - [x] 設定管理（internal/config/config.go）
  - [x] Lambdaハンドラ実装（cmd/lambda/main.go）
  - [x] 処理フロー統合
  - [x] ビルド確認（make build成功）

- [x] (2025-12-27) Milestone 8: CI/CD整備
  - [x] GitHub Actions設定（.github/workflows/ci.yml）
  - [x] lint + test + buildパイプライン

- [x] (2025-12-27) Milestone 9: インフラ構築
  - [x] Terraform初期化（terraform init成功）
  - [x] S3バケット定義（terraform/s3.tf）
  - [x] Lambda関数定義（terraform/lambda.tf）
  - [x] IAMロール・ポリシー定義（terraform/iam.tf）
  - [x] EventBridgeスケジュール定義（terraform/eventbridge.tf）
  - [x] terraform validate成功

- [x] (2025-12-27) Milestone 10: マルチインスタンス対応
  - [x] instance_name変数追加（terraform/variables.tf）
  - [x] locals.tf作成（name_prefix, common_tags定義）
  - [x] 全リソース名にlocal.name_prefixを反映
  - [x] IAMロール共通化（create_iam_role変数で制御）
  - [x] S3ポリシーをワイルドカード化（全インスタンス対応）
  - [x] タグ管理改善（Project, Instance, ManagedBy）
  - [x] terraform.tfvars.example更新
  - [x] terraform validate成功

- [ ] Milestone 11: デプロイと検証
  - [ ] 本番デプロイ
  - [ ] 動作確認
  - [ ] Discord通知受信確認

## Surprises & Discoveries

- Observation: golangci-lintのexportlooprefリンターがGo 1.22以降で非推奨になっている
  Evidence: `level=warning msg="The linter 'exportloopref' is deprecated (since v1.60.2) due to: Since Go1.22 (loopvar) this linter is no longer relevant. Replaced by copyloopvar."`
  対応: copyloopvarに置き換え

- Observation: go testの-raceフラグがGo 1.25.5（開発版）でcmd/lambdaパッケージに対してエラーを発生させる
  Evidence: `FAIL github.com/alp/suumo-hunter-go/cmd/lambda [setup failed]`
  対応: テスト対象をinternal/...に限定

## Decision Log

- Decision: golangci-lintのPATH設定をMakefileで明示的に行う
  Rationale: go installでインストールしたツールが$(GOPATH)/binにあるため、PATHに含まれていない環境でも動作するようにする
  Date/Author: 2025-12-27

- Decision: テスト対象をinternal/...に限定
  Rationale: mainパッケージ（cmd/lambda）のテストは現状不要であり、Go 1.25.5との互換性問題を回避するため
  Date/Author: 2025-12-27

## Outcomes & Retrospective

### Milestone 1 完了 (2025-12-27)

**達成事項:**
- Goプロジェクトの基本構造を構築
- SPEC.mdのセクション6に記載のディレクトリ構造を作成
- 必要な依存ライブラリをgo.modに追加（コードで使用時にgo mod tidyで反映）
- Makefile（build, lint, test, deploy）を作成
- golangci-lint設定を作成

**検証結果:**
- `make build` → ARM64バイナリ生成成功（build/bootstrap）
- `make lint` → エラーなし
- `make test` → 成功（テストファイルなしのためスキップ）

**次のステップ:**
Milestone 2でデータモデル（Property構造体）を実装する

### Milestone 2 完了 (2025-12-27)

**達成事項:**
- Property構造体を定義（ID, Name, Address, Age, Floor, Rent, ManagementFee, Deposit, KeyMoney, Layout, Area, WalkMinutes, URL）
- データ変換ユーティリティを実装:
  - `ParseRent()` - 「7.9万円」→ 79000.0
  - `ParseArea()` - 「25.5m²」→ 25.5
  - `ParseAge()` - 「築5年」→ 5、「新築」→ 0
  - `ParseWalkMinutes()` - 「歩8分」→ 8
  - `ParseFloor()` - 「3階」→ 3、「3-4階」→ 3
  - `ExtractPropertyID()` - URLから物件IDを抽出
- CSV読み書き機能を実装（LoadFromCSV, SaveToCSV）
- 差分検出機能（FindNewProperties）とマージ機能（MergeProperties）を実装
- ユニットテストを作成（カバレッジ90.2%）

**検証結果:**
- `make lint` → エラーなし
- `make test` → 全テストPASS

**次のステップ:**
Milestone 3でSUUMOスクレイピング機能を実装する

### Milestone 3 完了 (2025-12-27)

**達成事項:**
- SUUMOスクレイパーを実装（internal/scraper/suumo.go）
- goqueryを使用したHTMLパース処理
- ページネーション対応（最大ページ数設定可能）
- retry-goを使用したリトライ処理（指数バックオフ）
- 機能オプションパターンによる設定（WithMaxPages, WithRetryAttempts等）
- モックサーバーを使用した統合テスト
- ParseRent関数を改善（「5000円」形式もサポート）

**検証結果:**
- `make lint` → エラーなし
- `make test` → 全テストPASS（scraper: カバレッジ94.6%）

**次のステップ:**
Milestone 4でS3ストレージ機能を実装する

### Milestone 4 完了 (2025-12-27)

**達成事項:**
- S3ストレージクライアントを実装（internal/storage/s3.go）
- S3APIインターフェースを定義（モックテスト対応）
- CSVダウンロード機能（NoSuchKeyエラー時は空配列を返す）
- CSVアップロード機能（Content-Type設定）
- モックを使用した包括的なユニットテスト

**検証結果:**
- `make lint` → エラーなし
- `make test` → 全テストPASS（storage: カバレッジ86.2%）

**次のステップ:**
Milestone 5でDiscord通知機能を実装する

### Milestone 5 完了 (2025-12-27)

**達成事項:**
- Discord Webhookクライアントを実装（internal/notifier/discord.go）
- SPEC.md準拠のメッセージフォーマット
  - 総賃料（万円）、割安度（円/月）を表示
- 文字数制限・分割送信（1000文字上限、10件上限）
- HTTPClientインターフェースによるモック対応
- PropertyWithScore型の定義
- ConvertToPropertyWithScoreヘルパー関数

**検証結果:**
- `make lint` → エラーなし
- `make test` → 全テストPASS（notifier: カバレッジ92.6%）

**次のステップ:**
Milestone 6で重回帰分析機能を実装する

### Milestone 6 完了 (2025-12-27)

**達成事項:**
- 重回帰分析をgonum/matを使用して実装（internal/analyzer/regression.go）
- 目的変数: 総賃料（rent + management_fee）
- 説明変数: 専有面積、築年数、階数、駅徒歩分数
- 正規方程式による係数算出: β = (X'X)^(-1) X'y
- お得度（円）= 予測総賃料 - 実際総賃料
- 最低サンプル数: 10件（不足時は「分析中」ラベル）
- AnalyzeNewProperties: 全データで回帰、新着物件のみスコア算出

**検証結果:**
- `make lint` → エラーなし
- `make test` → 全テストPASS（analyzer: カバレッジ88.3%）

**次のステップ:**
Milestone 7でLambda統合を実装する


---


## Milestone 1: プロジェクト基盤構築

このマイルストーンでは、Goプロジェクトの基本構造を構築する。完了後、`go mod tidy` で依存関係が解決され、`make lint` でlintが実行できる状態になる。

### Context and Orientation

現在のワークスペースには以下のファイルが存在する：
- `config.json` - 設定ファイル（既存Python版の設定と思われる）
- `main.py` - Python版の実装
- `PLANS.md` - ExecPlan作成ガイドライン
- `SPEC.md` - システム仕様書

Goプロジェクトのファイルはまだ存在しない。SPEC.mdのセクション6に記載されたディレクトリ構造に従ってプロジェクトを構築する。

### Plan of Work

1. `go mod init` でGoモジュールを初期化する。モジュール名は `github.com/alp/suumo-hunter-go` とする。

2. 以下のディレクトリ構造を作成する：

        suumo-hunter-go/
        ├── cmd/
        │   └── lambda/
        ├── internal/
        │   ├── config/
        │   ├── scraper/
        │   ├── storage/
        │   ├── notifier/
        │   ├── analyzer/
        │   └── models/
        └── terraform/

3. `go.mod` に依存ライブラリを追加する：
   - github.com/aws/aws-lambda-go
   - github.com/aws/aws-sdk-go-v2
   - github.com/PuerkitoBio/goquery
   - github.com/avast/retry-go
   - github.com/caarlos0/env/v9
   - gonum.org/v1/gonum

4. `.golangci.yml` を作成し、Google Go Style Guideに準拠した設定を行う。

5. `Makefile` を作成し、以下のターゲットを定義する：
   - `build` - ARM64向けバイナリビルド
   - `lint` - golangci-lint実行
   - `test` - テスト実行
   - `deploy` - ビルド＋terraform apply

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # Go module初期化
    go mod init github.com/alp/suumo-hunter-go

    # 依存ライブラリ追加
    go get github.com/aws/aws-lambda-go@latest
    go get github.com/aws/aws-sdk-go-v2@latest
    go get github.com/aws/aws-sdk-go-v2/config@latest
    go get github.com/aws/aws-sdk-go-v2/service/s3@latest
    go get github.com/PuerkitoBio/goquery@latest
    go get github.com/avast/retry-go/v4@latest
    go get github.com/caarlos0/env/v9@latest
    go get gonum.org/v1/gonum@latest

    # lintツールインストール（未インストールの場合）
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

    # lint実行確認
    make lint

期待される出力（lint実行時）：

    $ make lint
    golangci-lint run
    # エラーなし、または軽微な警告のみ

### Validation and Acceptance

1. `go mod tidy` がエラーなく完了する
2. `make lint` がエラーなく完了する
3. SPEC.mdのセクション6に記載されたディレクトリ構造が作成されている

### Interfaces and Dependencies

この段階では空のパッケージのみ。各ディレクトリに `.gitkeep` または空の `doc.go` を配置してディレクトリを保持する。


---


## Milestone 2: データモデル実装

このマイルストーンでは、物件データを表す構造体とCSV読み書き機能を実装する。完了後、`go test ./internal/models/...` でテストが通過する。

### Context and Orientation

SPEC.md セクション4.1に定義された物件データフィールドをGoの構造体として実装する。

| フィールド | Go型 | 説明 |
|-----------|------|------|
| ID | string | 物件ID（jnc_XXXXXXXXXXXX形式） |
| Name | string | 物件名 |
| Address | string | 住所 |
| Age | int | 築年数 |
| Floor | int | 階数 |
| Rent | float64 | 家賃（円） |
| ManagementFee | float64 | 管理費（円） |
| Deposit | string | 敷金 |
| KeyMoney | string | 礼金 |
| Layout | string | 間取り |
| Area | float64 | 専有面積（m²） |
| WalkMinutes | int | 駅徒歩分数 |
| URL | string | 物件詳細URL |

SUUMOでは家賃が「7.9万円」のように表記されるため、パース時に10,000を乗じて円に変換する必要がある。

### Plan of Work

1. `internal/models/property.go` に `Property` 構造体を定義する

2. 以下のユーティリティ関数を実装する：
   - `ParseRent(s string) float64` - 「7.9万円」→ 79000.0
   - `ParseArea(s string) float64` - 「25.5m²」→ 25.5
   - `ParseAge(s string) int` - 「築5年」→ 5、「新築」→ 0
   - `ParseWalkMinutes(s string) int` - 「歩8分」→ 8

3. `internal/models/csv.go` にCSV読み書き機能を実装する：
   - `LoadFromCSV(r io.Reader) ([]Property, error)`
   - `SaveToCSV(w io.Writer, properties []Property) error`

4. `internal/models/property_test.go` にユニットテストを作成する

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # テスト実行
    go test -v ./internal/models/...

期待される出力：

    === RUN   TestParseRent
    --- PASS: TestParseRent (0.00s)
    === RUN   TestParseArea
    --- PASS: TestParseArea (0.00s)
    === RUN   TestParseAge
    --- PASS: TestParseAge (0.00s)
    === RUN   TestParseWalkMinutes
    --- PASS: TestParseWalkMinutes (0.00s)
    === RUN   TestCSVRoundTrip
    --- PASS: TestCSVRoundTrip (0.00s)
    PASS

### Validation and Acceptance

1. `go test ./internal/models/...` が全てPASS
2. 以下のパースが正しく動作することをテストで確認：
   - `ParseRent("7.9万円")` → `79000.0`
   - `ParseRent("10万円")` → `100000.0`
   - `ParseArea("25.5m²")` → `25.5`
   - `ParseAge("築5年")` → `5`
   - `ParseAge("新築")` → `0`
   - `ParseWalkMinutes("歩8分")` → `8`

### Interfaces and Dependencies

`internal/models/property.go` に定義する型と関数：

    package models

    type Property struct {
        ID            string
        Name          string
        Address       string
        Age           int
        Floor         int
        Rent          float64
        ManagementFee float64
        Deposit       string
        KeyMoney      string
        Layout        string
        Area          float64
        WalkMinutes   int
        URL           string
    }

    // TotalRent returns rent + management fee
    func (p Property) TotalRent() float64

    // ParseRent converts "7.9万円" to 79000.0
    func ParseRent(s string) (float64, error)

    // ParseArea converts "25.5m²" to 25.5
    func ParseArea(s string) (float64, error)

    // ParseAge converts "築5年" to 5, "新築" to 0
    func ParseAge(s string) (int, error)

    // ParseWalkMinutes converts "歩8分" to 8
    func ParseWalkMinutes(s string) (int, error)

`internal/models/csv.go` に定義する関数：

    // LoadFromCSV reads properties from CSV
    func LoadFromCSV(r io.Reader) ([]Property, error)

    // SaveToCSV writes properties to CSV
    func SaveToCSV(w io.Writer, properties []Property) error


---


## Milestone 3: スクレイピング機能実装

このマイルストーンでは、SUUMOの物件一覧ページからデータを取得する機能を実装する。完了後、`go test ./internal/scraper/...` でテストが通過し、実際のSUUMOページをスクレイピングできる。

### Context and Orientation

SUUMOの賃貸物件一覧ページは以下のような構造を持つ：
- 検索結果はページネーションされ、URLの `page` パラメータで制御
- 各物件は `div.cassetteitem` 内に格納
- 物件名、住所、築年数などが子要素として配置
- 物件詳細URLから物件ID（jnc_XXXXXXXXXXXX）を抽出

SPEC.md セクション5.1に記載の通り、スクレイピングは以下の要件を満たす：
- リトライ: 3回、10秒間隔、指数バックオフ
- 最大ページ数: 30（環境変数で変更可能）

### Plan of Work

1. `internal/scraper/suumo.go` にスクレイパーを実装する：
   - `Scraper` 構造体（HTTPクライアント、設定を保持）
   - `Scrape(ctx context.Context, baseURL string) ([]Property, error)`
   - ページネーション処理
   - リトライ処理（retry-go使用）

2. HTMLパースロジックを実装する：
   - `parsePropertyList(doc *goquery.Document) []Property`
   - `parseProperty(s *goquery.Selection) Property`

3. `internal/scraper/suumo_test.go` にユニットテストを作成する：
   - サンプルHTMLを用いたパーステスト
   - エッジケーステスト

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # テスト実行
    go test -v ./internal/scraper/...

    # 統合テスト（実際のSUUMOへのアクセス、オプション）
    go test -v -tags=integration ./internal/scraper/...

期待される出力：

    === RUN   TestParsePropertyList
    --- PASS: TestParsePropertyList (0.00s)
    === RUN   TestParseProperty
    --- PASS: TestParseProperty (0.00s)
    PASS

### Validation and Acceptance

1. `go test ./internal/scraper/...` が全てPASS
2. サンプルHTMLから物件データが正しくパースされることを確認
3. 以下のデータが正しく抽出される：
   - 物件名
   - 住所
   - 築年数（数値化）
   - 階数（数値化）
   - 家賃（円変換済み）
   - 管理費（円変換済み）
   - 敷金・礼金
   - 間取り
   - 専有面積（数値化）
   - 駅徒歩分数（最初の駅のみ）
   - 物件URL
   - 物件ID

### Interfaces and Dependencies

`internal/scraper/suumo.go` に定義する型と関数：

    package scraper

    type Scraper struct {
        client   *http.Client
        maxPages int
    }

    func NewScraper(maxPages int) *Scraper

    // Scrape fetches property listings from SUUMO
    func (s *Scraper) Scrape(ctx context.Context, baseURL string) ([]models.Property, error)


---


## Milestone 4: ストレージ機能実装

このマイルストーンでは、AWS S3との連携機能を実装する。完了後、`go test ./internal/storage/...` でテストが通過する。

### Context and Orientation

物件データはCSV形式でS3バケットに保存される。SPEC.md セクション4.4に記載の通り：
- 形式: CSV
- 保存先: AWS S3
- 重複排除キー: id（物件ID）

AWS SDK for Go v2を使用する。

### Plan of Work

1. `internal/storage/s3.go` にS3クライアントを実装する：
   - `Storage` 構造体
   - `Download(ctx context.Context) ([]Property, error)`
   - `Upload(ctx context.Context, properties []Property) error`

2. `internal/storage/s3_test.go` にユニットテストを作成する：
   - モックS3クライアントを使用

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # テスト実行
    go test -v ./internal/storage/...

期待される出力：

    === RUN   TestDownload
    --- PASS: TestDownload (0.00s)
    === RUN   TestUpload
    --- PASS: TestUpload (0.00s)
    PASS

### Validation and Acceptance

1. `go test ./internal/storage/...` が全てPASS
2. S3クライアントがモックと正しく連携することを確認

### Interfaces and Dependencies

`internal/storage/s3.go` に定義する型と関数：

    package storage

    type Storage struct {
        client     *s3.Client
        bucketName string
        bucketKey  string
    }

    func NewStorage(client *s3.Client, bucketName, bucketKey string) *Storage

    // Download fetches properties CSV from S3
    func (s *Storage) Download(ctx context.Context) ([]models.Property, error)

    // Upload saves properties CSV to S3
    func (s *Storage) Upload(ctx context.Context, properties []models.Property) error


---


## Milestone 5: 通知機能実装

このマイルストーンでは、Discord Webhook APIを使用した通知機能を実装する。完了後、`go test ./internal/notifier/...` でテストが通過する。

### Context and Orientation

SPEC.md セクション4.3およびセクション10.2に記載の通り：
- エンドポイント: Discord Webhook URL
- 認証: Bearer Token
- メソッド: POST
- Content-Type: application/x-www-form-urlencoded

通知フォーマット：

    🏠 新着物件のお知らせ

    🔥【お買い得】マンション名A
    📍 東京都渋谷区...
    💰 8.5万円（管理費込）
    💴 相場より 12,800円/月 お得！
    🔗 https://suumo.jp/...

制限事項：
- 1回の通知上限: 10件
- メッセージ長制限: 1000文字を超える場合は分割送信

### Plan of Work

1. `internal/notifier/discord.go` にDiscord Webhookクライアントを実装する：
   - `Notifier` 構造体
   - `Notify(ctx context.Context, properties []PropertyWithScore) error`
   - メッセージフォーマット生成
   - 文字数制限・分割送信ロジック

2. `PropertyWithScore` 型を定義する（割安度情報を含む）

3. `internal/notifier/discord_test.go` にユニットテストを作成する

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # テスト実行
    go test -v ./internal/notifier/...

期待される出力：

    === RUN   TestFormatMessage
    --- PASS: TestFormatMessage (0.00s)
    === RUN   TestSplitMessages
    --- PASS: TestSplitMessages (0.00s)
    PASS

### Validation and Acceptance

1. `go test ./internal/notifier/...` が全てPASS
2. メッセージフォーマットがSPEC.mdの仕様と一致することを確認
3. 10件超過時の要約メッセージが正しく生成されることを確認
4. 1000文字超過時の分割送信が正しく動作することを確認

### Interfaces and Dependencies

`internal/notifier/discord.go` に定義する型と関数：

    package notifier

    type PropertyWithScore struct {
        Property   models.Property
        Score      float64  // お得度（円）
        ScoreLabel string   // "お買い得", "標準", "割高", "分析中"
    }

    type Notifier struct {
        token string
    }

    func NewNotifier(token string) *Notifier

    // Notify sends Discord notification for new properties
    func (n *Notifier) Notify(ctx context.Context, properties []PropertyWithScore) error


---


## Milestone 6: 分析機能実装

このマイルストーンでは、重回帰分析による割安度判定機能を実装する。完了後、`go test ./internal/analyzer/...` でテストが通過する。

### Context and Orientation

SPEC.md セクション4.2に記載の重回帰分析仕様：

目的変数：総賃料（rent + management_fee）

説明変数：
- 専有面積（area）
- 築年数（age）
- 階数（floor）
- 駅徒歩分数（walk_minutes）

お得度の算出：

    お得度（円） = 予測総賃料 - 実際総賃料

判定基準：
- +10,000円以上: お買い得
- -10,000円〜+10,000円: 標準
- -10,000円未満: 割高

前提条件：
- 最低サンプル数: 10件以上
- サンプル不足時: 回帰分析をスキップ

gonum/stat パッケージを使用して重回帰分析を実装する。

### Plan of Work

1. `internal/analyzer/regression.go` に重回帰分析を実装する：
   - `Analyzer` 構造体
   - `Analyze(properties []Property) ([]PropertyWithScore, error)`
   - 重回帰係数の算出
   - 予測値の計算
   - お得度の算出

2. `internal/analyzer/regression_test.go` にユニットテストを作成する

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # テスト実行
    go test -v ./internal/analyzer/...

期待される出力：

    === RUN   TestAnalyze
    --- PASS: TestAnalyze (0.00s)
    === RUN   TestAnalyzeInsufficientSamples
    --- PASS: TestAnalyzeInsufficientSamples (0.00s)
    PASS

### Validation and Acceptance

1. `go test ./internal/analyzer/...` が全てPASS
2. 10件未満のサンプルでは分析がスキップされることを確認
3. お得度の判定基準が正しく適用されることを確認

### Interfaces and Dependencies

`internal/analyzer/regression.go` に定義する型と関数：

    package analyzer

    type Analyzer struct {
        minSamples int
    }

    func NewAnalyzer() *Analyzer

    // Analyze performs multiple regression analysis and calculates bargain scores
    func (a *Analyzer) Analyze(properties []models.Property) ([]notifier.PropertyWithScore, error)


---


## Milestone 7: Lambda統合

このマイルストーンでは、これまでの機能を統合してLambdaハンドラを実装する。完了後、ローカルでテスト実行ができる。

### Context and Orientation

SPEC.md セクション3の処理フロー：
1. EventBridgeが定期的にLambdaを起動
2. S3から前回取得した物件データ（CSV）をダウンロード
3. SUUMOの検索結果ページをスクレイピング（最大30ページ）
4. 前回データと比較して差分（新着物件）を検出
5. 取得データに対して重回帰分析を実行し、割安度を算出
6. 新しい物件データをS3にアップロード（CSV形式）
7. Discord Webhookで新着物件を通知（割安度付き）

環境変数（SPEC.md セクション7.1）：
- BUCKET_NAME（必須）
- BUCKET_KEY（デフォルト: properties.csv）
- MAX_PAGE（デフォルト: 30）
- SUUMO_SEARCH_URL（必須）
- DISCORD_WEBHOOK_URL（必須）

### Plan of Work

1. `internal/config/config.go` に設定管理を実装する：
   - `Config` 構造体
   - `Load() (*Config, error)` - 環境変数からロード

2. `cmd/lambda/main.go` にLambdaハンドラを実装する：
   - `Handler(ctx context.Context) error`
   - 処理フロー統合

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # ビルド確認
    make build

    # ローカルテスト（環境変数設定が必要）
    export BUCKET_NAME=your-bucket
    export SUUMO_SEARCH_URL="https://suumo.jp/..."
    export DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
    go run cmd/lambda/main.go

期待される出力（ローカル実行時）：

    Starting SUUMO Hunter...
    Downloading previous data from S3...
    Scraping SUUMO (max 30 pages)...
    Found 150 properties
    New properties: 5
    Running regression analysis...
    Uploading data to S3...
    Sending Discord notification...
    Done!

### Validation and Acceptance

1. `make build` がエラーなく完了し、ARM64バイナリが生成される
2. 環境変数を設定してローカル実行ができる
3. 処理フローが正しく動作する

### Interfaces and Dependencies

`internal/config/config.go` に定義する型：

    package config

    type Config struct {
        BucketName      string `env:"BUCKET_NAME,required"`
        BucketKey       string `env:"BUCKET_KEY" envDefault:"properties.csv"`
        MaxPage         int    `env:"MAX_PAGE" envDefault:"30"`
        SuumoSearchURL  string `env:"SUUMO_SEARCH_URL,required"`
        DiscordWebhookURL string `env:"DISCORD_WEBHOOK_URL,required"`
    }

    func Load() (*Config, error)


---


## Milestone 8: CI/CD整備

このマイルストーンでは、GitHub Actionsによる継続的インテグレーションを設定する。完了後、プルリクエスト時にlintとtestが自動実行される。

### Context and Orientation

SPEC.md セクション5.5に記載のCI要件：
- GitHub Actionsでパイプライン構築
- プルリクエスト時:
  - `golangci-lint run`
  - `go test ./...`

### Plan of Work

1. `.github/workflows/ci.yml` にワークフローを定義する：
   - Go 1.22セットアップ
   - 依存関係キャッシュ
   - golangci-lint実行
   - テスト実行

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # ローカルでCI相当の処理を実行して確認
    make lint
    make test

### Validation and Acceptance

1. `make lint` がエラーなく完了
2. `make test` が全てPASS
3. GitHubにpush後、CIが自動実行される

### Artifacts and Notes

`.github/workflows/ci.yml` の構造：

    name: CI
    on:
      pull_request:
      push:
        branches: [main]
    jobs:
      lint:
        runs-on: ubuntu-latest
        steps:
          - uses: actions/checkout@v4
          - uses: actions/setup-go@v5
            with:
              go-version: '1.22'
          - uses: golangci/golangci-lint-action@v4
      test:
        runs-on: ubuntu-latest
        steps:
          - uses: actions/checkout@v4
          - uses: actions/setup-go@v5
            with:
              go-version: '1.22'
          - run: go test ./...


---


## Milestone 9: インフラ構築

このマイルストーンでは、Terraformを使用してAWSインフラを構築する。完了後、`terraform plan` でリソース作成計画が確認できる。

### Context and Orientation

SPEC.md セクション2のアーキテクチャ図に基づき、以下のAWSリソースを構築する：
- S3バケット（CSVデータ保存）
- Lambda関数（Go ARM64）
- IAMロール・ポリシー
- EventBridge（定期実行）
- CloudWatch Logs（ログ出力）

### Plan of Work

1. `terraform/main.tf` - プロバイダー設定
2. `terraform/variables.tf` - 変数定義
3. `terraform/outputs.tf` - 出力定義
4. `terraform/s3.tf` - S3バケット
5. `terraform/lambda.tf` - Lambda関数
6. `terraform/iam.tf` - IAMロール・ポリシー
7. `terraform/eventbridge.tf` - 定期実行スケジュール

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go/terraform`

    # Terraform初期化
    terraform init

    # プラン確認
    terraform plan

期待される出力：

    Terraform will perform the following actions:

      # aws_iam_role.lambda will be created
      # aws_iam_role_policy.lambda will be created
      # aws_lambda_function.suumo_hunter will be created
      # aws_s3_bucket.properties will be created
      # aws_cloudwatch_event_rule.schedule will be created
      # aws_cloudwatch_event_target.lambda will be created

    Plan: 6 to add, 0 to change, 0 to destroy.

### Validation and Acceptance

1. `terraform init` が成功
2. `terraform plan` がエラーなく完了し、想定通りのリソースが表示される
3. `terraform validate` が成功


---


## Milestone 10: デプロイと検証

このマイルストーンでは、本番環境にデプロイし、実際にDiscord通知が届くことを確認する。

### Context and Orientation

Makefileの `deploy` ターゲットを使用してデプロイする。デプロイ後、手動でLambdaを実行して動作確認を行う。

### Plan of Work

1. Discord Webhook URLを取得し、環境変数またはterraform変数として設定
2. `make deploy` でビルドとデプロイを実行
3. AWS ConsoleまたはCLIからLambdaを手動実行
4. Discord通知が届くことを確認

### Concrete Steps

作業ディレクトリ: `/Users/alp/Projects/alp-dot/suumo-hunter-go`

    # DISCORD_WEBHOOK_URLを設定（Discordサーバー設定 > 連携サービス > ウェブフック で取得）
    export TF_VAR_discord_webhook_url="https://discord.com/api/webhooks/..."

    # デプロイ
    make deploy

    # 手動実行（AWS CLI）
    aws lambda invoke --function-name suumo-hunter output.json

    # 結果確認
    cat output.json

期待される出力（Lambda実行成功時）：

    {
        "StatusCode": 200,
        "ExecutedVersion": "$LATEST"
    }

Discord通知が届き、以下のような内容が表示される：

    🏠 新着物件のお知らせ

    🔥【お買い得】マンション名A
    📍 東京都渋谷区...
    ...

### Validation and Acceptance

1. Lambda関数が正常にデプロイされる
2. 手動実行でエラーが発生しない
3. Discord通知が正しいフォーマットで届く
4. S3にCSVファイルが保存される
5. CloudWatch Logsにログが出力される

### Idempotence and Recovery

- デプロイは何度実行しても同じ結果になる（Terraformの冪等性）
- Lambda実行が失敗した場合、CloudWatch Logsでエラー原因を確認
- S3ファイルが破損した場合、手動で削除して再実行すれば初期状態から再開
