/**
 * Framework（コンプライアンスフレームワーク）ドメインモデル
 * シンプルなEnumとエンティティの例
 */

import {
  Result,
  ok,
  err,
  ValidationError,
  validationError,
} from "../shared/result.js";
import { FrameworkId, ControlId, frameworkId } from "../shared/value-objects.js";

/**
 * FrameworkType: コンプライアンスフレームワークの種類
 * TypeScriptではstring literal unionで表現
 */
export type FrameworkType =
  | "SOC2"
  | "ISO27001"
  | "HIPAA"
  | "PCI_DSS"
  | "GDPR";

/**
 * FrameworkStatus: フレームワークの状態
 */
export type FrameworkStatus = "Draft" | "Active" | "Deprecated";

/**
 * Framework エンティティ
 */
export type Framework = {
  readonly id: FrameworkId;
  readonly type: FrameworkType;
  readonly name: string;
  readonly version: string;
  readonly description: string;
  readonly status: FrameworkStatus;
  readonly controlIds: readonly ControlId[];
};

export type CreateFrameworkInput = {
  readonly id: string;
  readonly type: FrameworkType;
  readonly name: string;
  readonly version: string;
  readonly description: string;
};

/**
 * Framework ファクトリ
 */
export const createFramework = (
  input: CreateFrameworkInput
): Result<Framework, ValidationError[]> => {
  const errors: ValidationError[] = [];

  // IDバリデーション
  const idResult = frameworkId(input.id);
  if (!idResult.ok) {
    errors.push(idResult.error);
  }

  // 名前バリデーション
  if (!input.name || input.name.trim().length === 0) {
    errors.push(
      validationError("name", "Framework name is required", "REQUIRED")
    );
  }

  // バージョンバリデーション（セマンティックバージョニング）
  const versionPattern = /^\d+\.\d+(\.\d+)?$/;
  if (!versionPattern.test(input.version)) {
    errors.push(
      validationError(
        "version",
        "Version must be in semver format (e.g., 1.0 or 1.0.0)",
        "INVALID_VERSION"
      )
    );
  }

  if (errors.length > 0) {
    return err(errors);
  }

  return ok({
    id: (idResult as { ok: true; value: FrameworkId }).value,
    type: input.type,
    name: input.name,
    version: input.version,
    description: input.description,
    status: "Draft" as FrameworkStatus,
    controlIds: [],
  });
};

/**
 * フレームワークのステータス更新
 */
export const updateFrameworkStatus = (
  framework: Framework,
  newStatus: FrameworkStatus
): Result<Framework, ValidationError> => {
  // ビジネスルール: Deprecated からは Active に戻せない
  if (framework.status === "Deprecated" && newStatus === "Active") {
    return err(
      validationError(
        "status",
        "Cannot reactivate a deprecated framework",
        "INVALID_TRANSITION"
      )
    );
  }

  // ビジネスルール: Controlが0件のFrameworkはActiveにできない
  if (newStatus === "Active" && framework.controlIds.length === 0) {
    return err(
      validationError(
        "status",
        "Cannot activate a framework without controls",
        "NO_CONTROLS"
      )
    );
  }

  return ok({
    ...framework,
    status: newStatus,
  });
};

/**
 * フレームワークにコントロールを追加
 */
export const addControlToFramework = (
  framework: Framework,
  controlId: ControlId
): Framework => {
  // 重複チェック
  if (framework.controlIds.includes(controlId)) {
    return framework;
  }

  return {
    ...framework,
    controlIds: [...framework.controlIds, controlId],
  };
};

/**
 * フレームワークタイプの表示用文字列
 */
export const getFrameworkTypeLabel = (type: FrameworkType): string => {
  switch (type) {
    case "SOC2":
      return "SOC 2";
    case "ISO27001":
      return "ISO 27001";
    case "HIPAA":
      return "HIPAA";
    case "PCI_DSS":
      return "PCI DSS";
    case "GDPR":
      return "GDPR";
  }
};
