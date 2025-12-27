package domain

import shared.*
import java.time.Instant

/**
 * RiskLevel: リスクの重大度レベル
 */
enum class RiskLevel(val value: Int) {
    Low(1),
    Medium(2),
    High(3),
    Critical(4)
}

/**
 * RiskScore: 不変のValue Object
 * value classでゼロコスト抽象化
 */
data class RiskScore private constructor(
    val likelihood: RiskLevel,
    val impact: RiskLevel,
    val value: Int,
    val label: String
) {
    companion object {
        fun calculate(likelihood: RiskLevel, impact: RiskLevel): RiskScore {
            val value = likelihood.value * impact.value
            val label = when {
                value <= 2 -> "Low"
                value <= 6 -> "Medium"
                value <= 12 -> "High"
                else -> "Critical"
            }
            return RiskScore(likelihood, impact, value, label)
        }
    }
}

/**
 * RiskCategory: リスクのカテゴリ
 */
enum class RiskCategory {
    Operational, Technical, Compliance, Financial
}

/**
 * RiskStatus: sealed classによるADT
 */
sealed class RiskStatus {
    data class Identified(val identifiedAt: Instant) : RiskStatus()

    data class Assessed(
        val assessedAt: Instant,
        val assessorId: UserId
    ) : RiskStatus()

    data class Mitigated(
        val mitigatedAt: Instant,
        val controlIds: List<ControlId>
    ) : RiskStatus()

    data class Accepted(
        val acceptedById: UserId,
        val reason: String,
        val expiresAt: Instant
    ) : RiskStatus()

    data class Closed(
        val closedAt: Instant,
        val resolution: String
    ) : RiskStatus()
}

/**
 * RiskStatusのパターンマッチ
 */
fun <T> RiskStatus.match(
    onIdentified: (Instant) -> T,
    onAssessed: (Instant, UserId) -> T,
    onMitigated: (Instant, List<ControlId>) -> T,
    onAccepted: (UserId, String, Instant) -> T,
    onClosed: (Instant, String) -> T
): T = when (this) {
    is RiskStatus.Identified -> onIdentified(identifiedAt)
    is RiskStatus.Assessed -> onAssessed(assessedAt, assessorId)
    is RiskStatus.Mitigated -> onMitigated(mitigatedAt, controlIds)
    is RiskStatus.Accepted -> onAccepted(acceptedById, reason, expiresAt)
    is RiskStatus.Closed -> onClosed(closedAt, resolution)
}

/**
 * Risk エンティティ
 */
data class Risk private constructor(
    val id: RiskId,
    val title: String,
    val description: String,
    val category: RiskCategory,
    val inherentScore: RiskScore,
    val residualScore: RiskScore,
    val status: RiskStatus,
    val ownerId: UserId
) {
    companion object {
        fun create(
            id: String,
            title: String,
            description: String,
            category: RiskCategory,
            likelihood: RiskLevel,
            impact: RiskLevel,
            ownerId: UserId
        ): Result<Risk, ValidationErrors> {
            val errors = mutableListOf<ValidationError>()

            val riskIdResult = RiskId.create(id)
            if (riskIdResult is Result.Err) {
                errors.add(riskIdResult.error)
            }

            if (title.isBlank()) {
                errors.add(ValidationError("title", "Risk title is required", "REQUIRED"))
            }

            if (errors.isNotEmpty()) {
                return err(ValidationErrors(errors))
            }

            val inherentScore = RiskScore.calculate(likelihood, impact)

            return ok(Risk(
                id = (riskIdResult as Result.Ok).value,
                title = title,
                description = description,
                category = category,
                inherentScore = inherentScore,
                residualScore = inherentScore, // 初期状態では同じ
                status = RiskStatus.Identified(Instant.now()),
                ownerId = ownerId
            ))
        }
    }

    /**
     * ステータス更新（バリデーション付き）
     */
    fun updateStatus(newStatus: RiskStatus): Result<Risk, ValidationError> {
        // Business rule: Closed状態からは遷移不可
        if (status is RiskStatus.Closed) {
            return err(ValidationError(
                "status",
                "Cannot transition from Closed status",
                "INVALID_TRANSITION"
            ))
        }

        // Business rule: Acceptedの有効期限チェック
        if (newStatus is RiskStatus.Accepted && newStatus.expiresAt <= Instant.now()) {
            return err(ValidationError(
                "expiresAt",
                "Acceptance expiration date must be in the future",
                "INVALID_EXPIRATION"
            ))
        }

        return ok(copy(status = newStatus))
    }

    /**
     * 残余リスクスコアの更新
     */
    fun updateResidualScore(likelihood: RiskLevel, impact: RiskLevel): Risk =
        copy(residualScore = RiskScore.calculate(likelihood, impact))
}

/**
 * リスクステータスの表示用文字列
 */
fun RiskStatus.toLabel(): String = match(
    onIdentified = { identifiedAt -> "特定済み ($identifiedAt)" },
    onAssessed = { assessedAt, _ -> "評価済み ($assessedAt)" },
    onMitigated = { _, controlIds -> "軽減済み (${controlIds.size}件の統制)" },
    onAccepted = { _, reason, expiresAt -> "受容 ($reason, 期限: $expiresAt)" },
    onClosed = { _, resolution -> "クローズ ($resolution)" }
)
