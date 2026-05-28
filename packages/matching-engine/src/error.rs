use thiserror::Error;

#[derive(Debug, Error)]
pub enum EngineError {
    #[error("invalid order: {0}")]
    InvalidOrder(String),

    #[error("order not found: {0}")]
    OrderNotFound(String),

    #[error("insufficient quantity")]
    InsufficientQuantity,

    #[error("internal error: {0}")]
    InternalError(String),
}
