package domain

import shared.*
import java.time.Instant

/**
 * ControlStatus: sealed classによるADT
 * Kotlinのsealed classはコンパイル時に網羅性をチェック可能
 */
sealed class ControlStatus {
    data object NotImplemented : ControlStatus()

    data class InProgress(val progress: Percentage) : ControlStatus()

    data class Implemented(val implementedAt: Instant) : ControlStatus()

    data class NotApplicable(val reason: String) : ControlStatus()

    data class Failed(
        val reason: String,
        val detectedAt: Instant
    ) : ControlStatus()
}

/**
 * when式によるパターンマッチ
 * sealed classなのでelse不要（コンパイル時に網羅性保証）
 */
fun <T> ControlStatus.match(
    onNotImplemented: () -> T,
    onInProgress: (Percentage) -> T,
    onImplemented: (Instant) -> T,
    onNotApplicable: (String) -> T,
    onFailed: (String, Instant) -> T
): T = when (this) {
    is ControlStatus.NotImplemented -> onNotImplemented()
    is ControlStatus.InProgress -> onInProgress(progress)
    is ControlStatus.Implemented -> onImplemented(implementedAt)
    is ControlStatus.NotApplicable -> onNotApplicable(reason)
    is ControlStatus.Failed -> onFailed(reason, detectedAt)
}

/**
 * Control エンティティ
 * data classで不変性を保証（valのみ使用）
 */
data class Control private constructor(
    val id: ControlId,
    val frameworkId: FrameworkId,
    val code: String,
    val title: String,
    val description: String,
    val status: ControlStatus,
    val ownerId: UserId
) {
    companion object {
        fun create(
            id: String,
            frameworkId: FrameworkId,
            code: String,
            title: String,
            description: String,
            ownerId: UserId
        ): Result<Control, ValidationErrors> {
            val errors = mutableListOf<ValidationError>()

            val controlIdResult = ControlId.create(id)
            if (controlIdResult is Result.Err) {
                errors.add(controlIdResult.error)
            }

            if (code.isBlank()) {
                errors.add(ValidationError("code", "Control code is required", "REQUIRED"))
            }

            if (title.isBlank()) {
                errors.add(ValidationError("title", "Control title is required", "REQUIRED"))
            }

            if (errors.isNotEmpty()) {
                return err(ValidationErrors(errors))
            }

            return ok(Control(
                id = (controlIdResult as Result.Ok).value,
                frameworkId = frameworkId,
                code = code,
                title = title,
                description = description,
                status = ControlStatus.NotImplemented,
                ownerId = ownerId
            ))
        }
    }

    /**
     * copy()で不変性を保ちながら状態を更新
     * ビジネスルールのバリデーション付き
     */
    fun updateStatus(newStatus: ControlStatus): Result<Control, ValidationError> {
        // Business rule: Failed状態から直接Implementedには遷移できない
        if (status is ControlStatus.Failed && newStatus is ControlStatus.Implemented) {
            return err(ValidationError(
                "status",
                "Cannot transition directly from Failed to Implemented",
                "INVALID_TRANSITION"
            ))
        }

        return ok(copy(status = newStatus))
    }
}

/**
 * ステータスの表示用文字列
 */
fun ControlStatus.toLabel(): String = match(
    onNotImplemented = { "未実装" },
    onInProgress = { progress -> "実装中 (${progress.value}%)" },
    onImplemented = { implementedAt -> "実装済み ($implementedAt)" },
    onNotApplicable = { reason -> "適用外: $reason" },
    onFailed = { reason, _ -> "失敗: $reason" }
)
