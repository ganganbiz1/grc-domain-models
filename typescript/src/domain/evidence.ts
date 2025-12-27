/**
 * Evidence（エビデンス）ドメインモデル
 * ADT検証: EvidenceType（各エビデンスタイプで異なるデータ構造）
 * バリデーション検証: 有効期限チェック
 */

import {
  Result,
  ok,
  err,
  ValidationError,
  validationError,
} from "../shared/result.js";
import {
  EvidenceId,
  ControlId,
  UserId,
  IntegrationId,
  Url,
  evidenceId,
  url,
} from "../shared/value-objects.js";

/**
 * FileType: ドキュメントのファイル形式
 */
export type FileType = "PDF" | "DOCX" | "XLSX" | "PNG" | "JPG";

/**
 * CheckResult: 自動チェックの結果
 */
export type CheckResult =
  | { readonly type: "Passed" }
  | { readonly type: "Failed"; readonly reason: string }
  | { readonly type: "Skipped"; readonly reason: string };

/**
 * EvidenceType: Discriminated Union
 * 各エビデンスタイプで保持するデータが異なる
 */
export type EvidenceType =
  | {
      readonly type: "Document";
      readonly fileUrl: Url;
      readonly fileType: FileType;
    }
  | {
      readonly type: "Screenshot";
      readonly imageUrl: Url;
      readonly capturedAt: Date;
    }
  | {
      readonly type: "AutomatedCheck";
      readonly integrationId: IntegrationId;
      readonly checkName: string;
      readonly lastRunAt: Date;
      readonly result: CheckResult;
    }
  | {
      readonly type: "ManualReview";
      readonly reviewerId: UserId;
      readonly reviewedAt: Date;
      readonly notes: string;
    };

// EvidenceType コンストラクタ
export const EvidenceType = {
  document: (fileUrl: Url, fileType: FileType): EvidenceType => ({
    type: "Document",
    fileUrl,
    fileType,
  }),

  screenshot: (imageUrl: Url, capturedAt: Date): EvidenceType => ({
    type: "Screenshot",
    imageUrl,
    capturedAt,
  }),

  automatedCheck: (
    integrationId: IntegrationId,
    checkName: string,
    lastRunAt: Date,
    result: CheckResult
  ): EvidenceType => ({
    type: "AutomatedCheck",
    integrationId,
    checkName,
    lastRunAt,
    result,
  }),

  manualReview: (
    reviewerId: UserId,
    reviewedAt: Date,
    notes: string
  ): EvidenceType => ({
    type: "ManualReview",
    reviewerId,
    reviewedAt,
    notes,
  }),
} as const;

/**
 * EvidenceStatus: エビデンスの状態
 */
export type EvidenceStatus = "Valid" | "Expired" | "Pending" | "Rejected";

/**
 * Evidence エンティティ
 */
export type Evidence = {
  readonly id: EvidenceId;
  readonly controlId: ControlId;
  readonly evidenceType: EvidenceType;
  readonly collectedAt: Date;
  readonly expiresAt: Date | null;
  readonly description: string;
};

export type CreateEvidenceInput = {
  readonly id: string;
  readonly controlId: ControlId;
  readonly evidenceType: EvidenceType;
  readonly collectedAt: Date;
  readonly expiresAt: Date | null;
  readonly description: string;
};

/**
 * Evidence ファクトリ: バリデーション付き
 */
export const createEvidence = (
  input: CreateEvidenceInput
): Result<Evidence, ValidationError[]> => {
  const errors: ValidationError[] = [];

  // IDバリデーション
  const idResult = evidenceId(input.id);
  if (!idResult.ok) {
    errors.push(idResult.error);
  }

  // 有効期限バリデーション: 過去の日付は設定不可（新規作成時）
  if (input.expiresAt !== null && input.expiresAt <= new Date()) {
    errors.push(
      validationError(
        "expiresAt",
        "Expiration date must be in the future",
        "INVALID_EXPIRATION"
      )
    );
  }

  // 収集日バリデーション: 未来の日付は不可
  if (input.collectedAt > new Date()) {
    errors.push(
      validationError(
        "collectedAt",
        "Collection date cannot be in the future",
        "INVALID_COLLECTION_DATE"
      )
    );
  }

  if (errors.length > 0) {
    return err(errors);
  }

  return ok({
    id: (idResult as { ok: true; value: EvidenceId }).value,
    controlId: input.controlId,
    evidenceType: input.evidenceType,
    collectedAt: input.collectedAt,
    expiresAt: input.expiresAt,
    description: input.description,
  });
};

/**
 * エビデンスの現在ステータスを計算
 * ドメインロジック: 有効期限が過ぎていればExpired
 */
export const getEvidenceStatus = (evidence: Evidence): EvidenceStatus => {
  if (evidence.expiresAt !== null && evidence.expiresAt < new Date()) {
    return "Expired";
  }

  // 自動チェックの場合は結果を確認
  if (evidence.evidenceType.type === "AutomatedCheck") {
    const result = evidence.evidenceType.result;
    if (result.type === "Failed") {
      return "Rejected";
    }
    if (result.type === "Skipped") {
      return "Pending";
    }
  }

  return "Valid";
};

/**
 * EvidenceType のパターンマッチ
 */
export const matchEvidenceType = <T>(
  evidenceType: EvidenceType,
  handlers: {
    onDocument: (fileUrl: Url, fileType: FileType) => T;
    onScreenshot: (imageUrl: Url, capturedAt: Date) => T;
    onAutomatedCheck: (
      integrationId: IntegrationId,
      checkName: string,
      lastRunAt: Date,
      result: CheckResult
    ) => T;
    onManualReview: (reviewerId: UserId, reviewedAt: Date, notes: string) => T;
  }
): T => {
  switch (evidenceType.type) {
    case "Document":
      return handlers.onDocument(evidenceType.fileUrl, evidenceType.fileType);
    case "Screenshot":
      return handlers.onScreenshot(
        evidenceType.imageUrl,
        evidenceType.capturedAt
      );
    case "AutomatedCheck":
      return handlers.onAutomatedCheck(
        evidenceType.integrationId,
        evidenceType.checkName,
        evidenceType.lastRunAt,
        evidenceType.result
      );
    case "ManualReview":
      return handlers.onManualReview(
        evidenceType.reviewerId,
        evidenceType.reviewedAt,
        evidenceType.notes
      );
    default:
      const _exhaustive: never = evidenceType;
      return _exhaustive;
  }
};

/**
 * エビデンスタイプの表示用文字列
 */
export const getEvidenceTypeLabel = (evidenceType: EvidenceType): string =>
  matchEvidenceType(evidenceType, {
    onDocument: (_url, fileType) => `ドキュメント (${fileType})`,
    onScreenshot: (_url, capturedAt) =>
      `スクリーンショット (${capturedAt.toISOString()})`,
    onAutomatedCheck: (_id, checkName, _date, result) =>
      `自動チェック: ${checkName} (${result.type})`,
    onManualReview: (_id, reviewedAt, _notes) =>
      `手動レビュー (${reviewedAt.toISOString()})`,
  });
