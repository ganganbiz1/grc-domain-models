package domain

import shared.*
import java.time.Instant

/**
 * FileType: ドキュメントのファイル形式
 */
enum class FileType {
    PDF, DOCX, XLSX, PNG, JPG
}

/**
 * CheckResult: 自動チェックの結果（sealed class）
 */
sealed class CheckResult {
    data object Passed : CheckResult()
    data class Failed(val reason: String) : CheckResult()
    data class Skipped(val reason: String) : CheckResult()
}

/**
 * EvidenceType: sealed classによるADT
 * 各エビデンスタイプで異なるデータ構造を持つ
 */
sealed class EvidenceType {
    data class Document(
        val fileUrl: Url,
        val fileType: FileType
    ) : EvidenceType()

    data class Screenshot(
        val imageUrl: Url,
        val capturedAt: Instant
    ) : EvidenceType()

    data class AutomatedCheck(
        val integrationId: IntegrationId,
        val checkName: String,
        val lastRunAt: Instant,
        val result: CheckResult
    ) : EvidenceType()

    data class ManualReview(
        val reviewerId: UserId,
        val reviewedAt: Instant,
        val notes: String
    ) : EvidenceType()
}

/**
 * EvidenceTypeのパターンマッチ
 */
fun <T> EvidenceType.match(
    onDocument: (Url, FileType) -> T,
    onScreenshot: (Url, Instant) -> T,
    onAutomatedCheck: (IntegrationId, String, Instant, CheckResult) -> T,
    onManualReview: (UserId, Instant, String) -> T
): T = when (this) {
    is EvidenceType.Document -> onDocument(fileUrl, fileType)
    is EvidenceType.Screenshot -> onScreenshot(imageUrl, capturedAt)
    is EvidenceType.AutomatedCheck -> onAutomatedCheck(integrationId, checkName, lastRunAt, result)
    is EvidenceType.ManualReview -> onManualReview(reviewerId, reviewedAt, notes)
}

/**
 * EvidenceStatus: エビデンスの状態
 */
enum class EvidenceStatus {
    Valid, Expired, Pending, Rejected
}

/**
 * Evidence エンティティ
 */
data class Evidence private constructor(
    val id: EvidenceId,
    val controlId: ControlId,
    val evidenceType: EvidenceType,
    val collectedAt: Instant,
    val expiresAt: Instant?,
    val description: String
) {
    companion object {
        fun create(
            id: String,
            controlId: ControlId,
            evidenceType: EvidenceType,
            collectedAt: Instant,
            expiresAt: Instant?,
            description: String
        ): Result<Evidence, ValidationErrors> {
            val errors = mutableListOf<ValidationError>()
            val now = Instant.now()

            val evidenceIdResult = EvidenceId.create(id)
            if (evidenceIdResult is Result.Err) {
                errors.add(evidenceIdResult.error)
            }

            // 有効期限バリデーション
            if (expiresAt != null && expiresAt <= now) {
                errors.add(ValidationError(
                    "expiresAt",
                    "Expiration date must be in the future",
                    "INVALID_EXPIRATION"
                ))
            }

            // 収集日バリデーション
            if (collectedAt > now) {
                errors.add(ValidationError(
                    "collectedAt",
                    "Collection date cannot be in the future",
                    "INVALID_COLLECTION_DATE"
                ))
            }

            if (errors.isNotEmpty()) {
                return err(ValidationErrors(errors))
            }

            return ok(Evidence(
                id = (evidenceIdResult as Result.Ok).value,
                controlId = controlId,
                evidenceType = evidenceType,
                collectedAt = collectedAt,
                expiresAt = expiresAt,
                description = description
            ))
        }
    }

    /**
     * 現在のステータスを計算
     */
    val status: EvidenceStatus
        get() {
            val now = Instant.now()

            // 有効期限チェック
            if (expiresAt != null && expiresAt < now) {
                return EvidenceStatus.Expired
            }

            // 自動チェックの結果を確認
            if (evidenceType is EvidenceType.AutomatedCheck) {
                return when (evidenceType.result) {
                    is CheckResult.Failed -> EvidenceStatus.Rejected
                    is CheckResult.Skipped -> EvidenceStatus.Pending
                    is CheckResult.Passed -> EvidenceStatus.Valid
                }
            }

            return EvidenceStatus.Valid
        }
}

/**
 * エビデンスタイプの表示用文字列
 */
fun EvidenceType.toLabel(): String = match(
    onDocument = { _, fileType -> "ドキュメント ($fileType)" },
    onScreenshot = { _, capturedAt -> "スクリーンショット ($capturedAt)" },
    onAutomatedCheck = { _, checkName, _, result ->
        val resultStr = when (result) {
            is CheckResult.Passed -> "Passed"
            is CheckResult.Failed -> "Failed"
            is CheckResult.Skipped -> "Skipped"
        }
        "自動チェック: $checkName ($resultStr)"
    },
    onManualReview = { _, reviewedAt, _ -> "手動レビュー ($reviewedAt)" }
)
