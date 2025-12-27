package shared

import java.net.URI

/**
 * Value Objects: 不変の値オブジェクト群
 * Kotlinではvalue class (inline class)で型安全なラッパーを作成
 */

// ID型: value classでゼロコスト抽象化
@JvmInline
value class FrameworkId private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<FrameworkId, ValidationError> =
            if (value.isBlank()) {
                err(ValidationError("id", "FrameworkId cannot be empty", "EMPTY_ID"))
            } else {
                ok(FrameworkId(value))
            }
    }
}

@JvmInline
value class ControlId private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<ControlId, ValidationError> =
            if (value.isBlank()) {
                err(ValidationError("id", "ControlId cannot be empty", "EMPTY_ID"))
            } else {
                ok(ControlId(value))
            }
    }
}

@JvmInline
value class EvidenceId private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<EvidenceId, ValidationError> =
            if (value.isBlank()) {
                err(ValidationError("id", "EvidenceId cannot be empty", "EMPTY_ID"))
            } else {
                ok(EvidenceId(value))
            }
    }
}

@JvmInline
value class RiskId private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<RiskId, ValidationError> =
            if (value.isBlank()) {
                err(ValidationError("id", "RiskId cannot be empty", "EMPTY_ID"))
            } else {
                ok(RiskId(value))
            }
    }
}

@JvmInline
value class UserId private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<UserId, ValidationError> =
            if (value.isBlank()) {
                err(ValidationError("id", "UserId cannot be empty", "EMPTY_ID"))
            } else {
                ok(UserId(value))
            }
    }
}

@JvmInline
value class IntegrationId private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<IntegrationId, ValidationError> =
            if (value.isBlank()) {
                err(ValidationError("id", "IntegrationId cannot be empty", "EMPTY_ID"))
            } else {
                ok(IntegrationId(value))
            }
    }
}

/**
 * Percentage: 0-100の範囲を保証する値オブジェクト
 */
@JvmInline
value class Percentage private constructor(val value: Int) {
    companion object {
        fun create(value: Int): Result<Percentage, ValidationError> =
            if (value < 0 || value > 100) {
                err(ValidationError(
                    "percentage",
                    "Percentage must be between 0 and 100",
                    "INVALID_PERCENTAGE"
                ))
            } else {
                ok(Percentage(value))
            }
    }
}

/**
 * URL値オブジェクト
 */
@JvmInline
value class Url private constructor(val value: String) {
    companion object {
        fun create(value: String): Result<Url, ValidationError> =
            try {
                URI(value).toURL()
                ok(Url(value))
            } catch (e: Exception) {
                err(ValidationError("url", "Invalid URL format", "INVALID_URL"))
            }
    }
}
