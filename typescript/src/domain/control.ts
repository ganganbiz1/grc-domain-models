/**
 * Control（統制）ドメインモデル
 * ADT検証の主要対象: ControlStatus
 */

import {
  Result,
  ok,
  err,
  ValidationError,
  validationError,
} from "../shared/result.js";
import {
  ControlId,
  FrameworkId,
  UserId,
  Percentage,
  controlId,
  percentage,
} from "../shared/value-objects.js";

/**
 * ControlStatus: Discriminated Union (Tagged Union) によるADT
 *
 * TypeScriptでは `type` プロパティをdiscriminatorとして使用し、
 * switch文で網羅性チェックが可能
 */
export type ControlStatus =
  | { readonly type: "NotImplemented" }
  | { readonly type: "InProgress"; readonly progress: Percentage }
  | { readonly type: "Implemented"; readonly implementedAt: Date }
  | { readonly type: "NotApplicable"; readonly reason: string }
  | {
      readonly type: "Failed";
      readonly reason: string;
      readonly detectedAt: Date;
    };

// ControlStatus コンストラクタ
export const ControlStatus = {
  notImplemented: (): ControlStatus => ({ type: "NotImplemented" }),

  inProgress: (progress: Percentage): ControlStatus => ({
    type: "InProgress",
    progress,
  }),

  implemented: (implementedAt: Date): ControlStatus => ({
    type: "Implemented",
    implementedAt,
  }),

  notApplicable: (reason: string): ControlStatus => ({
    type: "NotApplicable",
    reason,
  }),

  failed: (reason: string, detectedAt: Date): ControlStatus => ({
    type: "Failed",
    reason,
    detectedAt,
  }),
} as const;

/**
 * パターンマッチ: 網羅性チェック付き
 * TypeScriptでは never 型を使って網羅性を保証
 */
export const matchControlStatus = <T>(
  status: ControlStatus,
  handlers: {
    onNotImplemented: () => T;
    onInProgress: (progress: Percentage) => T;
    onImplemented: (implementedAt: Date) => T;
    onNotApplicable: (reason: string) => T;
    onFailed: (reason: string, detectedAt: Date) => T;
  }
): T => {
  switch (status.type) {
    case "NotImplemented":
      return handlers.onNotImplemented();
    case "InProgress":
      return handlers.onInProgress(status.progress);
    case "Implemented":
      return handlers.onImplemented(status.implementedAt);
    case "NotApplicable":
      return handlers.onNotApplicable(status.reason);
    case "Failed":
      return handlers.onFailed(status.reason, status.detectedAt);
    default:
      // 網羅性チェック: すべてのケースを処理していればここには到達しない
      const _exhaustive: never = status;
      return _exhaustive;
  }
};

/**
 * Control エンティティ
 * 不変性: readonly + private constructor パターン
 */
export type Control = {
  readonly id: ControlId;
  readonly frameworkId: FrameworkId;
  readonly code: string;
  readonly title: string;
  readonly description: string;
  readonly status: ControlStatus;
  readonly ownerId: UserId;
};

export type CreateControlInput = {
  readonly id: string;
  readonly frameworkId: FrameworkId;
  readonly code: string;
  readonly title: string;
  readonly description: string;
  readonly ownerId: UserId;
};

/**
 * Control ファクトリ: バリデーション付き
 */
export const createControl = (
  input: CreateControlInput
): Result<Control, ValidationError[]> => {
  const errors: ValidationError[] = [];

  // IDバリデーション
  const idResult = controlId(input.id);
  if (!idResult.ok) {
    errors.push(idResult.error);
  }

  // コードバリデーション
  if (!input.code || input.code.trim().length === 0) {
    errors.push(validationError("code", "Control code is required", "REQUIRED"));
  }

  // タイトルバリデーション
  if (!input.title || input.title.trim().length === 0) {
    errors.push(validationError("title", "Control title is required", "REQUIRED"));
  }

  if (errors.length > 0) {
    return err(errors);
  }

  return ok({
    id: (idResult as { ok: true; value: ControlId }).value,
    frameworkId: input.frameworkId,
    code: input.code,
    title: input.title,
    description: input.description,
    status: ControlStatus.notImplemented(),
    ownerId: input.ownerId,
  });
};

/**
 * 状態遷移: 不変性を保ちながら新しいControlを返す
 */
export const updateControlStatus = (
  control: Control,
  newStatus: ControlStatus
): Result<Control, ValidationError> => {
  // ビジネスルール: Failed状態からは直接Implementedに遷移できない
  if (
    control.status.type === "Failed" &&
    newStatus.type === "Implemented"
  ) {
    return err(
      validationError(
        "status",
        "Cannot transition directly from Failed to Implemented",
        "INVALID_TRANSITION"
      )
    );
  }

  return ok({
    ...control,
    status: newStatus,
  });
};

/**
 * ステータスの表示用文字列を取得
 */
export const getControlStatusLabel = (status: ControlStatus): string =>
  matchControlStatus(status, {
    onNotImplemented: () => "未実装",
    onInProgress: (progress) => `実装中 (${progress}%)`,
    onImplemented: (date) => `実装済み (${date.toISOString()})`,
    onNotApplicable: (reason) => `適用外: ${reason}`,
    onFailed: (reason, _date) => `失敗: ${reason}`,
  });
