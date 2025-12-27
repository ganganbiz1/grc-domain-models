/**
 * Result型: 成功または失敗を表す代数的データ型
 * 外部ライブラリを使わず、TypeScriptのDiscriminated Unionで実装
 */

export type Result<T, E> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: E };

// コンストラクタ関数
export const ok = <T>(value: T): Result<T, never> => ({
  ok: true,
  value,
});

export const err = <E>(error: E): Result<never, E> => ({
  ok: false,
  error,
});

// ユーティリティ関数
export const map = <T, U, E>(
  result: Result<T, E>,
  fn: (value: T) => U
): Result<U, E> => {
  if (result.ok) {
    return ok(fn(result.value));
  }
  return result;
};

export const flatMap = <T, U, E>(
  result: Result<T, E>,
  fn: (value: T) => Result<U, E>
): Result<U, E> => {
  if (result.ok) {
    return fn(result.value);
  }
  return result;
};

export const mapError = <T, E, F>(
  result: Result<T, E>,
  fn: (error: E) => F
): Result<T, F> => {
  if (!result.ok) {
    return err(fn(result.error));
  }
  return result;
};

export const getOrElse = <T, E>(result: Result<T, E>, defaultValue: T): T => {
  if (result.ok) {
    return result.value;
  }
  return defaultValue;
};

export const match = <T, E, U>(
  result: Result<T, E>,
  onOk: (value: T) => U,
  onErr: (error: E) => U
): U => {
  if (result.ok) {
    return onOk(result.value);
  }
  return onErr(result.error);
};

/**
 * ValidationError: ドメインバリデーションエラーの型
 */
export type ValidationError = {
  readonly field: string;
  readonly message: string;
  readonly code: string;
};

export const validationError = (
  field: string,
  message: string,
  code: string
): ValidationError => ({
  field,
  message,
  code,
});
