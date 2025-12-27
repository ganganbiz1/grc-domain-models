/**
 * Value Objects: 不変の値オブジェクト群
 * TypeScriptでは Branded Types パターンで型安全なIDを実現
 */

import { Result, ok, err, ValidationError, validationError } from "./result.js";

// Branded Types: コンパイル時の型安全性を保証
declare const brand: unique symbol;
type Brand<T, B> = T & { readonly [brand]: B };

// ID型
export type FrameworkId = Brand<string, "FrameworkId">;
export type ControlId = Brand<string, "ControlId">;
export type EvidenceId = Brand<string, "EvidenceId">;
export type RiskId = Brand<string, "RiskId">;
export type UserId = Brand<string, "UserId">;
export type IntegrationId = Brand<string, "IntegrationId">;

// IDファクトリ関数
const createId = <T extends string>(
  value: string,
  typeName: string
): Result<Brand<string, T>, ValidationError> => {
  if (!value || value.trim().length === 0) {
    return err(validationError("id", `${typeName} cannot be empty`, "EMPTY_ID"));
  }
  return ok(value as Brand<string, T>);
};

export const frameworkId = (value: string): Result<FrameworkId, ValidationError> =>
  createId(value, "FrameworkId");

export const controlId = (value: string): Result<ControlId, ValidationError> =>
  createId(value, "ControlId");

export const evidenceId = (value: string): Result<EvidenceId, ValidationError> =>
  createId(value, "EvidenceId");

export const riskId = (value: string): Result<RiskId, ValidationError> =>
  createId(value, "RiskId");

export const userId = (value: string): Result<UserId, ValidationError> =>
  createId(value, "UserId");

export const integrationId = (value: string): Result<IntegrationId, ValidationError> =>
  createId(value, "IntegrationId");

/**
 * Percentage: 0-100の範囲を保証する値オブジェクト
 */
export type Percentage = Brand<number, "Percentage">;

export const percentage = (value: number): Result<Percentage, ValidationError> => {
  if (value < 0 || value > 100) {
    return err(
      validationError(
        "percentage",
        "Percentage must be between 0 and 100",
        "INVALID_PERCENTAGE"
      )
    );
  }
  return ok(value as Percentage);
};

/**
 * URL値オブジェクト
 */
export type Url = Brand<string, "Url">;

export const url = (value: string): Result<Url, ValidationError> => {
  try {
    new URL(value);
    return ok(value as Url);
  } catch {
    return err(validationError("url", "Invalid URL format", "INVALID_URL"));
  }
};
