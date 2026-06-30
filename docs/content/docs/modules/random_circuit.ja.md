---
title: "random_circuit"
weight: 2
---

`x/random_circuit` モジュール（旧 `x/qcledger`）は、daqq のプロブレムシステムにおける**プロブレム #1** です。各ビーコンラウンドのシードから生成されたランダム量子回路について、参加者の**理論的出力確率分布**を記録します。

{{< callout type="info" >}}
**ケース A vs ケース B** — `x/random_circuit` は解析的／理論的分布（ケース A）を扱います。将来の `x/random_circuit_sampling` モジュールが、レジストリ内の別プロブレムとして経験的なショットヒストグラム（ケース B）を扱う予定です。
{{< /callout >}}

## x/problems への登録

`x/random_circuit` はチェーンジェネシス時に [problems](problems) レジストリに自己登録します。名前で検索：

```bash
quantumchaind query problems get-problem-by-name random_circuit
```

```yaml
problem:
  id: "1"
  name: random_circuit
  module_name: random_circuit
  kind: PROBLEM_KIND_BUILTIN
  enabled: true
  description: Theoretical output distribution of a randomly generated quantum circuit (case A).
```

keeper は割り当てられた `ProblemID` を独自のストアに永続化するため、他のモジュールやクライアントは再クエリせずにこれを導けます。

## beacon への依存

すべての提出は `roundID` を運びます。受理する前にハンドラは次を呼びます：

```go
// quantum-chain/x/random_circuit/keeper/msg_server_submit_result.go
_, err := k.beaconKeeper.GetSeed(ctx, msg.RoundId)
if err != nil {
    return nil, errorsmod.Wrapf(types.ErrSeedNotReady, "seed for round %d not ready", msg.RoundId)
}
```

これにより以下が保証されます：

1. 記録されるすべての結果が、すべてのノードがすでに合意しているシードに紐付いている。
2. 正直なノードがそのシードから同じ回路を生成すれば、同じ期待出力を出すため、食い違う結果が容易に発見できる。

## 提出形式

`MsgSubmitResult` は参加者の完全な理論分布を運びます：

```proto
message MsgSubmitResult {
  string creator = 1;
  uint64 round_id = 2;
  Distribution distribution = 3;
}

message Distribution {
  // 基底状態（例 "00101"）→ 十進文字列としての確率
  map<string, string> probabilities = 1;
}
```

確率は十進文字列としてエンコードされ、合意のシリアライズ中のノード間の浮動小数点の発散を避けます。

提出者は `(ラウンド, プロブレム)` のペアごとに最大 1 つの解を提出できます。再提出は拒否されます。

## TODO

このページにはまだ以下が必要です：

- `Distribution` の検証ルール（合計が約 1、次元 = `2^width`、バイト上限）。
- 格納された提出のクエリエンドポイント。
- 完全な提出フローの動作例。
