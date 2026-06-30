---
title: "クイックスタート"
weight: 1
---

すべてのコマンドはリポジトリのルートにいて、[Task](https://taskfile.dev/) がインストールされていることを前提とします。

## ノードバイナリのインストール

```bash
task -t Taskfile.quickstart.yml install
```

これは `quantumchaind` をコンパイルして `$GOPATH/bin` にインストールします。

## シングルノードチェーン

```bash
task -t Taskfile.quickstart.yml quickstart
```

`quickstart:init`（1 回だけ）と `quickstart:start` を実行するのと同等です。ホームディレクトリ：`~/.quantumchain`。

リセット：

```bash
task -t Taskfile.quickstart.yml quickstart:clean
```

## 共有乱数を生成する

beacon モジュールは、参加者の commit/reveal を元に 50 ブロックごとにネットワーク
共有の乱数 seed を生成します。単一ノードでも生成できます。ノードを起動した状態で
別ターミナルで beacon エージェントを走らせると、`alice` が毎ラウンド自動で commit と
reveal を行います。

```bash
task -t Taskfile.quickstart.yml beacon:loop
```

1 ラウンドは 50 ブロック（commit はオフセット 0〜30、reveal は 31〜45、確定は次の
境界）なので、最初の seed は数分で現れます。エージェントは `Ctrl+C` で停止します。

## ノードが動いているかを GUI で見てみよう

ノードには Web ビジュアライザが組み込まれており、ノード自身の REST API サーバが `/gui` で配信します。ネットワーク共有の乱数 seed（beacon の出力）を表示し、ノードのピアネットワークを図示します。

`quickstart:init` が REST API を有効化済みなので、ノードを起動した状態でビジュアライザをブラウザで開くだけです：

```bash
task -t Taskfile.quickstart.yml gui
```

これは `http://localhost:1317/gui/` を開きます。ページ・seed 用エンドポイント（`/gui/seeds`）・ネットワーク用エンドポイント（`/gui/net_info`）はすべてノード自身が同一オリジンで配信するため、CORS 設定も別途の Web サーバも一切不要です。接続に成功するとステータスバッジが緑色になり、上部パネルに最新の共有乱数 seed（と直近ラウンド）が、ネットワークパネルに自ノードと接続中のピアが表示されます。

## 3 ノードローカルネット（alice / bob / carol）

```bash
task -t Taskfile.quickstart.yml localnet:init
```

その後、各ノードをそれぞれのターミナルで起動：

```bash
task -t Taskfile.quickstart.yml localnet:start:alice
task -t Taskfile.quickstart.yml localnet:start:bob
task -t Taskfile.quickstart.yml localnet:start:carol
```

伝搬を確認：

```bash
task -t Taskfile.quickstart.yml localnet:status
```

リセット：

```bash
task -t Taskfile.quickstart.yml localnet:clean
```

## ランダム量子回路を実行（スタンドアロン）

```bash
task -t Taskfile.quickstart.yml circuit
```

タイムスタンプシードを使ってランダム量子回路を生成します。（このジェネレータをビーコンシードに配線するのは予定された統合です — [ビーコン → 統合](../modules/beacon#統合) を参照してください。）
