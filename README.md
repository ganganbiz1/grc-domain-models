# GRC Domain Models

VantaやDrataのようなコンプライアンスSaaS（GRC: Governance, Risk, Compliance）のドメインモデルを**TypeScript、Go、Kotlin**の3言語で実装し、表現力を比較するプロジェクト。

## 目的

プログラミング言語によるドメインモデルの表現力の差を検証し、技術選定の参考にする。

## 検証観点

| 観点 | 説明 |
|------|------|
| **ADT（代数的データ型）** | Union型、Enum、sealed classなどの表現力 |
| **バリデーション** | ドメインルールの検証とエラー処理 |
| **不変性** | データの不変性の保証しやすさ |

## ドメインモデル

| モデル | 説明 |
|--------|------|
| **Framework** | SOC2, ISO27001などのコンプライアンスフレームワーク |
| **Control** | 統制（フレームワークの要件を満たす施策） |
| **Evidence** | エビデンス（統制が実施されている証拠） |
| **Risk** | リスク（組織が直面するセキュリティリスク） |

---

## 言語比較

### 1. ADT（代数的データ型）

#### ControlStatus の実装比較

**TypeScript** - Discriminated Union
```typescript
type ControlStatus =
  | { readonly type: "NotImplemented" }
  | { readonly type: "InProgress"; readonly progress: Percentage }
  | { readonly type: "Implemented"; readonly implementedAt: Date }
  | { readonly type: "NotApplicable"; readonly reason: string }
  | { readonly type: "Failed"; readonly reason: string; readonly detectedAt: Date };
```

**Go** - Interface + 構造体（sealed pattern）
```go
type ControlStatus interface {
    controlStatus() // unexported method prevents external implementations
}

type NotImplemented struct{}
func (NotImplemented) controlStatus() {}

type InProgress struct { Progress Percentage }
func (InProgress) controlStatus() {}
```

**Kotlin** - sealed class
```kotlin
sealed class ControlStatus {
    data object NotImplemented : ControlStatus()
    data class InProgress(val progress: Percentage) : ControlStatus()
    data class Implemented(val implementedAt: Instant) : ControlStatus()
    data class NotApplicable(val reason: String) : ControlStatus()
    data class Failed(val reason: String, val detectedAt: Instant) : ControlStatus()
}
```

#### ADT比較表

| 観点 | TypeScript | Go | Kotlin |
|------|------------|-----|--------|
| 構文 | Discriminated Union | Interface + 構造体 | sealed class |
| 網羅性チェック | `never`型で実現 | なし（runtime panic） | コンパイル時 |
| データ保持 | オブジェクト | 構造体 | data class |
| パターンマッチ | switch文 | type switch | when式 |
| 記述量 | 中 | 多 | 少 |

### 2. バリデーション

#### Result型の実装比較

**TypeScript**
```typescript
type Result<T, E> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: E };

// 使用例
const createControl = (input): Result<Control, ValidationError[]> => {
  if (errors.length > 0) return err(errors);
  return ok(control);
};
```

**Go**
```go
type Result[T any] struct {
    value T
    err   error
    ok    bool
}

// 使用例（慣用的にはerror interfaceを使用）
func NewControl(input Input) (*Control, error) {
    if errors.HasErrors() {
        return nil, errors
    }
    return &Control{...}, nil
}
```

**Kotlin**
```kotlin
sealed class Result<out T, out E> {
    data class Ok<out T>(val value: T) : Result<T, Nothing>()
    data class Err<out E>(val error: E) : Result<Nothing, E>()
}

// 使用例
fun create(input: Input): Result<Control, ValidationErrors> {
    if (errors.isNotEmpty()) return err(ValidationErrors(errors))
    return ok(Control(...))
}
```

#### バリデーション比較表

| 観点 | TypeScript | Go | Kotlin |
|------|------------|-----|--------|
| エラー型 | Result<T, E> | error interface | Result<T, E> (sealed) |
| 複数エラー | ValidationError[] | ValidationErrors | ValidationErrors |
| 型安全性 | 高 | 中（interface） | 高 |
| 合成 | flatMap/chain | 手動 | flatMap |

### 3. 不変性（イミュータビリティ）

#### 不変性の実現方法

**TypeScript**
```typescript
// readonly + private constructor
type Control = {
  readonly id: ControlId;
  readonly status: ControlStatus;
  // ...
};

// 更新はspread演算子で新しいオブジェクトを作成
const updateStatus = (control: Control, status: ControlStatus): Control => ({
  ...control,
  status,
});
```

**Go**
```go
// unexportedフィールド + getterメソッド
type Control struct {
    id     ControlID  // unexported
    status ControlStatus
}

func (c *Control) ID() ControlID { return c.id }

// 更新は新しい構造体を作成
func (c *Control) WithStatus(status ControlStatus) *Control {
    return &Control{
        id:     c.id,
        status: status,
    }
}
```

**Kotlin**
```kotlin
// data class（valのみ使用）+ copy()
data class Control(
    val id: ControlId,
    val status: ControlStatus,
    // ...
) {
    fun updateStatus(newStatus: ControlStatus): Control =
        copy(status = newStatus)
}
```

#### 不変性比較表

| 観点 | TypeScript | Go | Kotlin |
|------|------------|-----|--------|
| 基本手法 | readonly + spread | unexported + getter | val + data class |
| 更新方法 | `{...obj, prop}` | 手動コピー | `copy()` |
| 強制力 | 型システム | 慣習的（破れる） | 型システム |
| 記述量 | 少 | 多 | 最少 |

---

## 総合評価

| 観点 | TypeScript | Go | Kotlin |
|------|------------|-----|--------|
| **ADT表現力** | ★★★★☆ | ★★☆☆☆ | ★★★★★ |
| **網羅性チェック** | ★★★★☆ | ★☆☆☆☆ | ★★★★★ |
| **バリデーション** | ★★★★☆ | ★★★☆☆ | ★★★★★ |
| **不変性** | ★★★★☆ | ★★★☆☆ | ★★★★★ |
| **記述の簡潔さ** | ★★★★☆ | ★★☆☆☆ | ★★★★★ |
| **学習コスト** | ★★★☆☆ | ★★★★★ | ★★★☆☆ |
| **エコシステム** | ★★★★★ | ★★★★☆ | ★★★★☆ |

### 言語別の特徴

#### TypeScript
- **強み**: Discriminated Unionで柔軟なADTを表現可能。既存のJavaScriptエコシステムとの親和性。
- **弱み**: ランタイムでは型情報が消えるため、完全な型安全性は保証されない。

#### Go
- **強み**: シンプルで学習コストが低い。パフォーマンスが高い。
- **弱み**: ADTの表現力が低く、パターンマッチの網羅性をコンパイル時にチェックできない。

#### Kotlin
- **強み**: sealed classによる最も表現力の高いADT。data classによる簡潔な不変性。whenの網羅性チェック。
- **弱み**: JVMに依存（ただしKotlin/Nativeもある）。

---

## ディレクトリ構成

```
grc-domain-models/
├── README.md
├── typescript/
│   ├── package.json
│   ├── tsconfig.json
│   └── src/
│       ├── shared/
│       │   ├── result.ts
│       │   └── value-objects.ts
│       └── domain/
│           ├── control.ts
│           ├── evidence.ts
│           ├── risk.ts
│           └── framework.ts
├── go/
│   ├── go.mod
│   └── domain/
│       ├── shared/
│       │   ├── result.go
│       │   └── valueobject.go
│       ├── control.go
│       ├── evidence.go
│       ├── risk.go
│       └── framework.go
└── kotlin/
    ├── build.gradle.kts
    ├── settings.gradle.kts
    └── src/main/kotlin/
        ├── shared/
        │   ├── Result.kt
        │   └── ValueObjects.kt
        └── domain/
            ├── Control.kt
            ├── Evidence.kt
            ├── Risk.kt
            └── Framework.kt
```

## セットアップ

### TypeScript
```bash
cd typescript
npm install
npm run typecheck
```

### Go
```bash
cd go
go build ./...
```

### Kotlin
```bash
cd kotlin
./gradlew build
```
