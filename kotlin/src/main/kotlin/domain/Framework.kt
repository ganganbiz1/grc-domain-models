package domain

import shared.*

/**
 * FrameworkType: コンプライアンスフレームワークの種類
 */
enum class FrameworkType(val displayName: String) {
    SOC2("SOC 2"),
    ISO27001("ISO 27001"),
    HIPAA("HIPAA"),
    PCI_DSS("PCI DSS"),
    GDPR("GDPR")
}

/**
 * FrameworkStatus: フレームワークの状態
 */
enum class FrameworkStatus {
    Draft, Active, Deprecated
}

/**
 * Framework エンティティ
 */
data class Framework private constructor(
    val id: FrameworkId,
    val type: FrameworkType,
    val name: String,
    val version: String,
    val description: String,
    val status: FrameworkStatus,
    val controlIds: List<ControlId>
) {
    companion object {
        private val semverPattern = Regex("""^\d+\.\d+(\.\d+)?$""")

        fun create(
            id: String,
            type: FrameworkType,
            name: String,
            version: String,
            description: String
        ): Result<Framework, ValidationErrors> {
            val errors = mutableListOf<ValidationError>()

            val frameworkIdResult = FrameworkId.create(id)
            if (frameworkIdResult is Result.Err) {
                errors.add(frameworkIdResult.error)
            }

            if (name.isBlank()) {
                errors.add(ValidationError("name", "Framework name is required", "REQUIRED"))
            }

            if (!semverPattern.matches(version)) {
                errors.add(ValidationError(
                    "version",
                    "Version must be in semver format (e.g., 1.0 or 1.0.0)",
                    "INVALID_VERSION"
                ))
            }

            if (errors.isNotEmpty()) {
                return err(ValidationErrors(errors))
            }

            return ok(Framework(
                id = (frameworkIdResult as Result.Ok).value,
                type = type,
                name = name,
                version = version,
                description = description,
                status = FrameworkStatus.Draft,
                controlIds = emptyList()
            ))
        }
    }

    /**
     * ステータス更新（バリデーション付き）
     */
    fun updateStatus(newStatus: FrameworkStatus): Result<Framework, ValidationError> {
        // Business rule: Deprecatedからはactiveに戻せない
        if (status == FrameworkStatus.Deprecated && newStatus == FrameworkStatus.Active) {
            return err(ValidationError(
                "status",
                "Cannot reactivate a deprecated framework",
                "INVALID_TRANSITION"
            ))
        }

        // Business rule: Controlが0件のFrameworkはActiveにできない
        if (newStatus == FrameworkStatus.Active && controlIds.isEmpty()) {
            return err(ValidationError(
                "status",
                "Cannot activate a framework without controls",
                "NO_CONTROLS"
            ))
        }

        return ok(copy(status = newStatus))
    }

    /**
     * コントロールを追加
     * 不変性を保ちながら新しいFrameworkを返す
     */
    fun addControl(controlId: ControlId): Framework {
        // 重複チェック
        if (controlId in controlIds) {
            return this
        }

        return copy(controlIds = controlIds + controlId)
    }
}
