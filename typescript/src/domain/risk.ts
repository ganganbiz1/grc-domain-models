/**
 * Risk（リスク）ドメインモデル
 * ADT検証: RiskStatus（状態遷移）
 * 不変性検証: RiskScore（Value Objectによる計算）
 */

import {
  Result,
  ok,
  err,
  ValidationError,
  validationError,
} from "../shared/result.js";
import { RiskId, ControlId, UserId, riskId } from "../shared/value-objects.js";

/**
 * RiskLevel: リスクの重大度レベル
 */
export type RiskLevel = "Low" | "Medium" | "High" | "Critical";

const riskLevelValue = (level: RiskLevel): number => {
  switch (level) {
    case "Low":
      return 1;
    case "Medium":
      return 2;
    case "High":
      return 3;
    case "Critical":
      return 4;
  }
};

/**
 * RiskScore: 不変のValue Object
 * likelihood × impact で計算される
 */
export type RiskScore = {
  readonly likelihood: RiskLevel;
  readonly impact: RiskLevel;
  readonly value: number;
  readonly label: string;
};

export const calculateRiskScore = (
  likelihood: RiskLevel,
  impact: RiskLevel
): RiskScore => {
  const value = riskLevelValue(likelihood) * riskLevelValue(impact);
  const label =
    value <= 2 ? "Low" : value <= 6 ? "Medium" : value <= 12 ? "High" : "Critical";
  return {
    likelihood,
    impact,
    value,
    label,
  };
};

/**
 * RiskCategory: リスクのカテゴリ
 */
export type RiskCategory = "Operational" | "Technical" | "Compliance" | "Financial";

/**
 * RiskStatus: Discriminated Union
 * 各状態で保持するデータが異なる
 */
export type RiskStatus =
  | { readonly type: "Identified"; readonly identifiedAt: Date }
  | {
      readonly type: "Assessed";
      readonly assessedAt: Date;
      readonly assessorId: UserId;
    }
  | {
      readonly type: "Mitigated";
      readonly mitigatedAt: Date;
      readonly controlIds: readonly ControlId[];
    }
  | {
      readonly type: "Accepted";
      readonly acceptedById: UserId;
      readonly reason: string;
      readonly expiresAt: Date;
    }
  | {
      readonly type: "Closed";
      readonly closedAt: Date;
      readonly resolution: string;
    };

// RiskStatus コンストラクタ
export const RiskStatus = {
  identified: (identifiedAt: Date): RiskStatus => ({
    type: "Identified",
    identifiedAt,
  }),

  assessed: (assessedAt: Date, assessorId: UserId): RiskStatus => ({
    type: "Assessed",
    assessedAt,
    assessorId,
  }),

  mitigated: (
    mitigatedAt: Date,
    controlIds: readonly ControlId[]
  ): RiskStatus => ({
    type: "Mitigated",
    mitigatedAt,
    controlIds,
  }),

  accepted: (
    acceptedById: UserId,
    reason: string,
    expiresAt: Date
  ): RiskStatus => ({
    type: "Accepted",
    acceptedById,
    reason,
    expiresAt,
  }),

  closed: (closedAt: Date, resolution: string): RiskStatus => ({
    type: "Closed",
    closedAt,
    resolution,
  }),
} as const;

/**
 * Risk エンティティ
 */
export type Risk = {
  readonly id: RiskId;
  readonly title: string;
  readonly description: string;
  readonly category: RiskCategory;
  readonly inherentScore: RiskScore;
  readonly residualScore: RiskScore;
  readonly status: RiskStatus;
  readonly ownerId: UserId;
};

export type CreateRiskInput = {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly category: RiskCategory;
  readonly likelihood: RiskLevel;
  readonly impact: RiskLevel;
  readonly ownerId: UserId;
};

/**
 * Risk ファクトリ: バリデーション付き
 */
export const createRisk = (
  input: CreateRiskInput
): Result<Risk, ValidationError[]> => {
  const errors: ValidationError[] = [];

  // IDバリデーション
  const idResult = riskId(input.id);
  if (!idResult.ok) {
    errors.push(idResult.error);
  }

  // タイトルバリデーション
  if (!input.title || input.title.trim().length === 0) {
    errors.push(validationError("title", "Risk title is required", "REQUIRED"));
  }

  if (errors.length > 0) {
    return err(errors);
  }

  const inherentScore = calculateRiskScore(input.likelihood, input.impact);

  return ok({
    id: (idResult as { ok: true; value: RiskId }).value,
    title: input.title,
    description: input.description,
    category: input.category,
    inherentScore,
    residualScore: inherentScore, // 初期状態では同じ
    status: RiskStatus.identified(new Date()),
    ownerId: input.ownerId,
  });
};

/**
 * リスクステータスの遷移: バリデーション付き
 */
export const updateRiskStatus = (
  risk: Risk,
  newStatus: RiskStatus
): Result<Risk, ValidationError> => {
  // ビジネスルール: Closed状態からは遷移不可
  if (risk.status.type === "Closed") {
    return err(
      validationError(
        "status",
        "Cannot transition from Closed status",
        "INVALID_TRANSITION"
      )
    );
  }

  // ビジネスルール: Accepted状態の有効期限チェック
  if (newStatus.type === "Accepted" && newStatus.expiresAt <= new Date()) {
    return err(
      validationError(
        "expiresAt",
        "Acceptance expiration date must be in the future",
        "INVALID_EXPIRATION"
      )
    );
  }

  return ok({
    ...risk,
    status: newStatus,
  });
};

/**
 * 残余リスクスコアの更新（軽減策適用後）
 */
export const updateResidualScore = (
  risk: Risk,
  likelihood: RiskLevel,
  impact: RiskLevel
): Risk => ({
  ...risk,
  residualScore: calculateRiskScore(likelihood, impact),
});

/**
 * RiskStatus のパターンマッチ
 */
export const matchRiskStatus = <T>(
  status: RiskStatus,
  handlers: {
    onIdentified: (identifiedAt: Date) => T;
    onAssessed: (assessedAt: Date, assessorId: UserId) => T;
    onMitigated: (mitigatedAt: Date, controlIds: readonly ControlId[]) => T;
    onAccepted: (acceptedById: UserId, reason: string, expiresAt: Date) => T;
    onClosed: (closedAt: Date, resolution: string) => T;
  }
): T => {
  switch (status.type) {
    case "Identified":
      return handlers.onIdentified(status.identifiedAt);
    case "Assessed":
      return handlers.onAssessed(status.assessedAt, status.assessorId);
    case "Mitigated":
      return handlers.onMitigated(status.mitigatedAt, status.controlIds);
    case "Accepted":
      return handlers.onAccepted(
        status.acceptedById,
        status.reason,
        status.expiresAt
      );
    case "Closed":
      return handlers.onClosed(status.closedAt, status.resolution);
    default:
      const _exhaustive: never = status;
      return _exhaustive;
  }
};

/**
 * リスクステータスの表示用文字列
 */
export const getRiskStatusLabel = (status: RiskStatus): string =>
  matchRiskStatus(status, {
    onIdentified: (date) => `特定済み (${date.toISOString()})`,
    onAssessed: (date, _assessorId) => `評価済み (${date.toISOString()})`,
    onMitigated: (date, controlIds) =>
      `軽減済み (${controlIds.length}件の統制)`,
    onAccepted: (_userId, reason, expiresAt) =>
      `受容 (${reason}, 期限: ${expiresAt.toISOString()})`,
    onClosed: (date, resolution) => `クローズ (${resolution})`,
  });
