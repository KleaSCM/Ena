# 🌸 Ena - あなたの優しい仮想アシスタント 🌸

Enaちゃんは、あなたのシステムを優しく管理してくれる仮想アシスタントです。Go言語で開発され、ファイル操作、ターミナル制御、アプリケーション管理、システム健康チェックなど、あらゆるシステム操作を美しく実行します。

## ✨ 特徴

- 🖥️ **包括的なシステム制御**: ファイル、フォルダ、ターミナル、アプリケーションを完全制御
- 🏥 **システム健康監視**: CPU、メモリ、ディスク使用状況をリアルタイム監視
- 🔍 **高度な検索機能**: ファイル検索と安全な削除機能
- ⚡ **システム操作**: 再起動、シャットダウン、スリープ機能
- 🎨 **美しいインターフェース**: カラフルで直感的なコマンドラインインターフェース
- 💕 **優しい日本語対応**: あたしの愛を込めた日本語メッセージ

## 🚀 インストール・実行

### 前提条件

- Go 1.21 以上
- Linux、macOS、またはWindows

### ビルド手順

```bash
# リポジトリをクローン
git clone <repository-url>
cd Ena

# 依存関係をインストール
go mod tidy

# ビルド
go build -o ena cmd/main.go
```

### 🎯 実行方法

#### インタラクティブモード（推奨）

```bash
# Enaちゃんを起動して対話モードを開始
./ena

# または
./ena --help  # ヘルプを表示
```

#### 直接コマンド実行

```bash
# システムの健康状態をチェック
./ena health

# ファイルを作成
./ena file create /path/to/file.txt

# アプリを起動
./ena app start firefox

# システム情報を表示
./ena system info
```

#### 実行例

```bash
# Enaちゃんを起動
$ ./ena

🌸 Enaちゃん - あなたの優しい仮想アシスタント 🌸
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
こんにちは！あたしはEnaよ〜 (๑˃̵ᴗ˂̵) あなたのお手伝いをさせていただくわ！

💡 ヒント: 'help' と入力すると、あたしができることを教えてあげる！
💡 ヒント: 'exit' と入力すると、あたしとお別れできるの...

Ena> health
🏥 システム健康診断レポート
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💻 CPU情報:
   モデル: AMD Ryzen 7 8845HS w/ Radeon 780M Graphics
   コア数: 16
   使用率: 8.8%
   状態: 🟢 正常
...

Ena> file create test.txt
ファイル「test.txt」を作成しました！ (๑˃̵ᴗ˂̵)

Ena> exit
Enaちゃん、お疲れ様でした〜 また会いましょうね！ (╹◡╹)♡
```

## 📖 使用方法

### 基本操作

Enaちゃんは2つの方法で使用できます：

1. **インタラクティブモード**: `./ena` で起動し、コマンドを対話的に入力
2. **直接実行**: `./ena <コマンド>` で特定のコマンドを直接実行

### ヘルプとサポート

```bash
# 全体的なヘルプを表示
./ena --help

# 特定のコマンドのヘルプを表示
./ena file --help
./ena app --help
./ena system --help
```

## 🎯 コマンド一覧

### 📁 ファイル操作

```bash
# ファイルを作成
ena file create /path/to/file.txt

# ファイルを読み込み
ena file read /path/to/file.txt

# ファイルに書き込み
ena file write /path/to/file.txt "Hello, World!"

# ファイルをコピー
ena file copy /source.txt /dest.txt

# ファイルを移動
ena file move /old.txt /new.txt

# ファイルを削除
ena file delete /path/to/file.txt

# ファイル情報を表示
ena file info /path/to/file.txt
```

### 📂 フォルダ操作

```bash
# フォルダを作成
ena folder create /path/to/folder

# フォルダ内容を一覧表示
ena folder list /path/to/folder

# フォルダを削除
ena folder delete /path/to/folder

# フォルダ情報を表示
ena folder info /path/to/folder
```

### 🖥️ ターミナル操作

```bash
# 新しいターミナルを開く
ena terminal open

# ターミナルを閉じる
ena terminal close

# コマンドを実行
ena terminal execute "ls -la"

# ディレクトリを変更
ena terminal cd /home/user
```

### 📱 アプリケーション操作

```bash
# アプリを起動
ena app start firefox

# アプリを停止
ena app stop firefox

# アプリを再起動
ena app restart firefox

# 起動中のアプリ一覧
ena app list

# アプリ情報を表示
ena app info firefox
```

### ⚡ システム操作

```bash
# システムを再起動
ena system restart

# システムをシャットダウン
ena system shutdown

# システムをスリープ
ena system sleep

# システム情報を表示
ena system info
```

### 🏥 システム健康チェック

```bash
# システムの健康状態をチェック
ena health
```

### 🔍 検索・削除

```bash
# ファイルを検索
ena search "*.txt" /home/user

# ファイルを削除
ena delete /path/to/file.txt
```

## 🏗️ アーキテクチャ

```
Ena/
├── cmd/ena/              # メインエントリーポイント
├── internal/             # 内部パッケージ
│   ├── core/            # コアエンジン
│   ├── hooks/           # システムフック
│   ├── health/          # システム健康監視
│   └── utils/           # ユーティリティ
├── pkg/                  # 公開パッケージ
│   ├── commands/        # コマンド定義
│   └── system/          # システム操作
├── Docs/                # ドキュメント
└── Tests/               # テストファイル
```

## 🛡️ 安全性

- **安全モード**: デフォルトで有効化されており、危険な操作前に確認を求めます
- **危険コマンド検出**: システムに害を与える可能性のあるコマンドを自動検出
- **エラーハンドリング**: 包括的なエラーハンドリングとユーザーフレンドリーなエラーメッセージ

## 🎨 カスタマイズ

Enaちゃんの外観や動作は、設定ファイルや環境変数でカスタマイズできます。

## 🔧 トラブルシューティング

### よくある問題

**Q: ビルド時にエラーが発生します**
```bash
# 依存関係を再インストール
go clean -modcache
go mod tidy
go build -o ena cmd/main.go
```

**Q: ターミナルが開きません**
- システムにインストールされているターミナルエミュレータを確認
- 対応: gnome-terminal, xterm, konsole, xfce4-terminal, alacritty, kitty

**Q: アプリが起動しません**
- アプリ名が正しいか確認（例: firefox, chrome, vim）
- アプリがシステムにインストールされているか確認

**Q: システム操作で権限エラーが発生します**
```bash
# sudo権限が必要な場合があります
sudo ./ena system restart
```

## 🤝 コントリビューション

Enaちゃんをより良くするためのコントリビューションを歓迎します！

1. リポジトリをフォーク
2. フィーチャーブランチを作成
3. 変更をコミット
4. プルリクエストを送信

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 💕 作者

**Author**: KleaSCM  
**Email**: KleaSCM@gmail.com

---

Enaちゃんと一緒に、あなたのコンピューターライフを楽しくしましょうね！ (╹◡╹)♡
