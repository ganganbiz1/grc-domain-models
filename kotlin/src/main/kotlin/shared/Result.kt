package shared

/**
 * Result型: 成功または失敗を表すsealed class
 * Kotlinのsealed classはADTを完璧に表現できる
 */
sealed class Result<out T, out E> {
    data class Ok<out T>(val value: T) : Result<T, Nothing>()
    data class Err<out E>(val error: E) : Result<Nothing, E>()

    val isOk: Boolean get() = this is Ok
    val isErr: Boolean get() = this is Err

    fun <U> map(fn: (T) -> U): Result<U, E> = when (this) {
        is Ok -> Ok(fn(value))
        is Err -> this
    }

    fun <U> flatMap(fn: (T) -> Result<U, @UnsafeVariance E>): Result<U, E> = when (this) {
        is Ok -> fn(value)
        is Err -> this
    }

    fun <F> mapError(fn: (E) -> F): Result<T, F> = when (this) {
        is Ok -> this
        is Err -> Err(fn(error))
    }

    fun getOrElse(default: @UnsafeVariance T): T = when (this) {
        is Ok -> value
        is Err -> default
    }

    fun <U> match(onOk: (T) -> U, onErr: (E) -> U): U = when (this) {
        is Ok -> onOk(value)
        is Err -> onErr(error)
    }
}

// 拡張関数でコンストラクタを提供
fun <T> ok(value: T): Result<T, Nothing> = Result.Ok(value)
fun <E> err(error: E): Result<Nothing, E> = Result.Err(error)

/**
 * ValidationError: ドメインバリデーションエラー
 */
data class ValidationError(
    val field: String,
    val message: String,
    val code: String
) {
    override fun toString(): String = "[$code] $field: $message"
}

/**
 * ValidationErrors: 複数のバリデーションエラーを保持
 * Nel (Non-Empty List) の簡易版
 */
data class ValidationErrors(
    val errors: List<ValidationError>
) {
    init {
        require(errors.isNotEmpty()) { "ValidationErrors must contain at least one error" }
    }

    val first: ValidationError get() = errors.first()

    fun add(error: ValidationError): ValidationErrors =
        ValidationErrors(errors + error)

    override fun toString(): String =
        errors.joinToString("; ") { it.toString() }

    companion object {
        fun of(error: ValidationError): ValidationErrors =
            ValidationErrors(listOf(error))

        fun of(vararg errors: ValidationError): ValidationErrors =
            ValidationErrors(errors.toList())
    }
}
