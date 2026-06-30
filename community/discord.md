# daqq Discord — セットアップ・ブループリント

「コミュニティ形成」（提案名: 「…DAQQ の開発と**コミュニティ形成**」）の一環としての公式 Discord。
サーバー作成自体は要ログイン操作なので人手で行い、文面・構成はこのドキュメントから流用する。
作成後の招待リンクは公開サイト（daqq.pages.dev）・ドキュメント・提出物に組み込む。

## サーバー名
**daqq — Distributed Agreement on Quantum Queries**

## チャンネル構成

```
📌 INFORMATION
  #welcome        … 自己紹介の前に読む入口
  #announcements  … 運営からの告知（読み取り専用）
  #rules          … ルール
  #links          … 公式リンク集

💬 COMMUNITY
  #general        … 雑談
  #introductions  … 自己紹介
  #ideas          … アイデア・提案

🔬 PROTOCOL & DEV
  #dev                … 実装の議論
  #beacon-circuits    … シード合意・ランダム回路
  #cross-validation   … クロス検証・結果の一致/相違
  #run-a-node         … ノード運用の質問
  #github             … （任意）GitHub 通知の自動投稿

🧪 QUANTUM
  #algorithms     … 量子アルゴリズム全般
  #papers         … 論文・参考資料

🛠 SUPPORT
  #help           … 困りごと
```

> 多言語化するなら `#general-ja` / `#general-en` を併設。最初は日本語1本で十分。

## ロール
- `@core`（運営）／`@node-operator`（ノード運用者）／`@contributor`（貢献者）／`@member`（既定）
- `@announcements`（告知メンション用・任意でオプトイン）

## #links に貼る公式リンク
- ライブダッシュボード: https://daqq.pages.dev/gui/
- ドキュメント: https://daqq.pages.dev/docs/
- トップ: https://daqq.pages.dev/
- GitHub: https://github.com/massn/daqq

---

## 貼り付け用テキスト

### #welcome
```
daqq コミュニティへようこそ 👋

daqq（Distributed Agreement on Quantum Queries）は、無報酬の分散台帳です。
P2Pノードが同じ共有ランダムシードに合意し、同一の量子回路を各自で計算して、
結果をオンチェーンに記録・相互検証します。「誰の量子計算結果でも、特定主体を
信頼せずに再現・監査できる」公共的な基盤を、みんなで運用するのが目的です。

▶ ライブダッシュボード: https://daqq.pages.dev/gui/
📖 ドキュメント: https://daqq.pages.dev/docs/

まずは #rules を読んで、#introductions で自己紹介をどうぞ。
量子・ブロックチェーン・分散システムに興味があれば大歓迎です。
```

### #rules
```
1. 敬意を持って接する。ハラスメント・差別・スパムは禁止。
2. 投機・価格・トークンセールの話題は対象外（daqqは無報酬・報酬トークン無し）。
3. 質問は該当チャンネルで。コードやログはコードブロックで共有。
4. 個人情報・ノードのIPアドレス等は公開しない。
5. 日本語・英語どちらでもOK。
```

### #introductions（テンプレ）
```
- 名前/ハンドル:
- 興味分野（量子 / ブロックチェーン / 分散システム など）:
- daqqに期待すること:
```

---

## 作成手順（5分）
1. Discordで「サーバーを追加 ＋」→「自分で作成」→「自分と友達のため」。
2. サーバー名を `daqq` に。
3. 上のカテゴリ／チャンネルを作成（最初は INFORMATION と COMMUNITY だけでも可）。
4. 各文面を貼り付け。
5. サーバー設定 → ロールで上記ロールを追加。
6. **招待リンクを発行**（招待 → Edit invite link → Expire after: **Never**, Max uses: **No limit**）。

---

## 招待リンク
**https://discord.gg/pxrjYJKKF** （server: daqq ／ 期限なし推奨）

## 組み込み状況
- [x] `daqq.pages.dev` トップ（`docs/content/_index.md` / `_index.ja.md`）の CTA に **💬 Discord** ボタン追加（要 Pages 再デプロイで本番反映）
- [ ] ドキュメントにコミュニティ導線
- [ ] 成果報告書／概要版に「コミュニティ: discord.gg/pxrjYJKKF」を一行（コミュニティ形成の実績として・任意）
- [ ] Discord サーバー #links の GitHub 行を最新に
