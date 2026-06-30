---
title: "daqq"
toc: false
---

<p style="text-align: center; font-size: 1.25rem; font-weight: 600; margin: 1.5rem 0;">共有された入力、共有された記録 — ブロックチェーン上で。</p>

**daqq**（Distributed Agreement on Quantum Queries）は、無報酬の分散台帳です。P2Pノードが**同じブロック高で同じ新鮮なランダム値に合意**し、それをシードとして**同一の量子アルゴリズム**を実行し、**各ノードの結果をオンチェーンに記録**することで、ネットワーク全体で比較・監査できるようにします。

これは決済ネットワークでも、スマートコントラクトのプラットフォームでも、特定アルゴリズムのベンチマークサービスでもありません。「すべてのノードが同じランダム入力を与えられたときに何が起きたか」の**共有された改ざん耐性のある記録**であり、量子ハードウェアとシミュレータのクロスバリデーション、再現可能なランダム化ベンチマーク、分散型の科学的記録管理に有用です。

<div style="display:flex; gap:0.75rem; justify-content:center; flex-wrap:wrap; margin:2rem 0;">
  <a href="/gui/" style="display:inline-block; padding:0.75rem 1.6rem; border-radius:10px; background:#3b82f6; color:#fff; font-weight:700; text-decoration:none;">▶ ライブダッシュボード（GUI）</a>
  <a href="/docs/" style="display:inline-block; padding:0.75rem 1.6rem; border-radius:10px; border:2px solid #3b82f6; color:#3b82f6; font-weight:700; text-decoration:none;">📖 ドキュメント</a>
  <a href="https://discord.gg/pxrjYJKKF" style="display:inline-block; padding:0.75rem 1.6rem; border-radius:10px; background:#5865F2; color:#fff; font-weight:700; text-decoration:none;">💬 Discord</a>
</div>

<p style="text-align:center; color:#6b7280; font-size:0.9rem; margin-top:-1rem;">ライブダッシュボードでは、共有シード・登録された問題・ノードネットワーク・独立バリデータによるラウンドごとのクロス検証が見られます。</p>

{{< cards >}}
  {{< card link="docs/overview" title="概要" subtitle="daqqとは何か、なぜ存在するのか。" >}}
  {{< card link="docs/concepts" title="コンセプト" subtitle="ブロック、コミット、リビール、ラウンド、シード、プロブレム。" >}}
  {{< card link="docs/architecture" title="アーキテクチャ" subtitle="Cosmos SDK のモジュールと実行順序。" >}}
  {{< card link="docs/problem-system" title="プロブレムシステム" subtitle="一つのビーコン上で複数のアルゴリズムが共存する仕組み。" >}}
  {{< card link="docs/modules/beacon" title="ビーコンプロトコル" subtitle="RANDAO 方式のコミット・リビールでシードを合意する。" >}}
  {{< card link="docs/limitations" title="既知の制約" subtitle="公平性、同時性、セキュリティに関する注意点。" >}}
  {{< card link="docs/operations/quickstart" title="クイックスタート" subtitle="シングルノードまたは3ノードのローカルネットを起動する。" >}}
{{< /cards >}}

## 主な特性

- **報酬トークンなし。** 参加者はインセンティブのためではなく、台帳そのものを共に運用することを目的とします。MEVも、手数料市場も、インフレスケジュールもありません。
- **共有ランダム性。** 50ブロックごとに、すべてのノードがRANDAO方式のコミット・リビールとXOR集約によって同じ256ビットのシードを導きます。少なくとも一人の正直な参加者が高エントロピーの秘密を寄与する限り、シードは事前に誰にも予測できません。
- **マルチプロブレム設計。** チェーンはプラットフォームです。各量子アルゴリズムは独自の Cosmos SDK モジュールに格納され、オンチェーンの `problems` レジストリに登録されます。新しいアルゴリズムは gov アップグレードによって新規モジュールとして追加されます。最初のものである `random_circuit` は、各ラウンドのシードからランダム回路を生成し、各参加者の理論的出力分布を記録します。
- **監査可能で再現可能。** 誰でもオンチェーンのリビールからシードを再導出でき、シードから各アルゴリズムの入力を再導出できます。ノード間の不一致は可視であり、再生可能です。
