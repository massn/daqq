---
title: "ドキュメント"
weight: 1
sidebar:
  open: true
---

daqq のドキュメントへようこそ。

- [概要]({{< relref "overview" >}}) — daqq とは何か、設計意図
- [コンセプト]({{< relref "concepts" >}}) — 用語（ブロック、コミット、リビール、ラウンド、シード）とラウンドのライフサイクル
- [アーキテクチャ]({{< relref "architecture" >}}) — モジュール構成と実行順序
- [プロブレムシステム]({{< relref "problem-system" >}}) — マルチプロブレム・フレームワークの設計仕様
- [既知の制約]({{< relref "limitations" >}}) — 未解決の問題、公平性に関する注意、深刻度の評価
- モジュール
  - [beacon]({{< relref "modules/beacon" >}}) — RANDAO コミット・リビール方式のランダムネスビーコン
  - [random_circuit]({{< relref "modules/random_circuit" >}}) — プロブレム #1：理論出力分布の台帳
  - [problems]({{< relref "modules/problems" >}}) — オンチェーンのプロブレムレジストリ
  - [quantumchain]({{< relref "modules/quantumchain" >}}) — ベースモジュール
- [運用]({{< relref "operations/quickstart" >}}) — クイックスタートとローカルネット
