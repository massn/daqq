---
title: "problems"
weight: 3
---

`x/problems` モジュールは daqq プロブレムのオンチェーン**レジストリ**です。それ自体はどのプロブレムも実装しません — どのプロブレムモジュールが存在し、それぞれに割り当てられた ID と、現在提出を受け付けているかを追跡します。

全体の設計の理由は [プロブレムシステム]({{< relref "problem-system" >}}) ページを参照してください。

## 何を格納するか

```go
type Problem struct {
    ID           uint64
    Name         string      // すべてのプロブレムで一意（例："random_circuit"）
    ModuleName   string      // 実装する Cosmos モジュール名
    Kind         ProblemKind // MVP では BUILTIN；WASM は将来のために予約
    Enabled      bool        // false なら新規提出を受け付けない、履歴は保持
    AddedAtRound uint64      // このプロブレムが選択可能になったビーコンラウンド
    Description  string
}

type Params struct {
    NextProblemID uint64      // 単調増加；ID は決して再利用されない
}
```

状態の不変条件：

- ID は順次割り当てられ、プロブレムが無効化された後も**決して再利用されません**。
- `Name` は無効化されたものを含めすべてのプロブレムで一意です。
- プロブレムは**追加または切り替えのみ**で、決して削除されません。

## プロブレムモジュールはどのように自己登録するか

各プロブレムモジュール（例：`x/random_circuit`）は自身の `InitGenesis` から `ProblemsKeeper.RegisterProblem(ctx, name, moduleName, description)` を呼びます。`Register` は `name` に対してべき等なので、ジェネシスを再実行（例：テスト中や `quickstart:clean` の後）しても安全です。

これを機能させるため、`app/app_config.go` の `InitGenesis` 順序ではプロブレムモジュールより**前**に `problems` を置いています：

```go
InitGenesis: []string{
    // ...
    problemsmoduletypes.ModuleName,
    randomcircuitmoduletypes.ModuleName,
}
```

## クエリ

| コマンド | 目的 |
|---|---|
| `quantumchaind query problems params` | `NextProblemID` を表示。 |
| `quantumchaind query problems list-problems` | 登録されているすべてのプロブレムを列挙（ページング付き）。 |
| `quantumchaind query problems get-problem [id]` | 数値 ID で検索。 |
| `quantumchaind query problems get-problem-by-name [name]` | 一意な名前で検索。 |

## プロブレムの無効化

無効化はエントリを削除しません — `enabled` フラグを切り替えるだけです。実装モジュールのメッセージハンドラは、プロブレムが無効化されているときに新規提出を拒否する責任を負います（`ProblemsKeeper.IsEnabled(ctx, id)`）。

切り替え自体は gov ゲート付きです：`MsgSetProblemEnabled` を通常のガバナンス提案として `x/gov` 経由で送ります。承認時：

```yaml
tx:
  body:
    messages:
      - "@type": /quantumchain.problems.v1.MsgSetProblemEnabled
        authority: "<gov module address>"
        id: 1
        enabled: false
```

今や無効化されたプロブレムへの履歴の提出はクエリ可能なまま残ります。

## 新しいプロブレムの追加

新しいプロブレムは次のように追加されます：

1. `BeaconKeeper.GetSeed` と `ProblemsKeeper.RegisterProblem` に依存する新しい `x/<your_problem>` モジュールを書く。
2. 新しいバイナリでリリースを切る。
3. リリースを指す `software-upgrade` gov 提案を提出し、ネットワークをロールする。
4. アップグレード高さで、新しいモジュールの `InitGenesis` / アップグレードハンドラが `RegisterProblem` を実行し、レジストリエントリが現れる。

[プロブレムシステム → 新しいプロブレムの追加]({{< relref "problem-system#新しいプロブレムの追加コミュニティ貢献" >}}) を参照してください。
