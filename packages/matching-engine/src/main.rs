use std::net::SocketAddr;

use tokio::net::TcpListener;
use tracing_subscriber::EnvFilter;

mod engine;
mod error;
mod orderbook;
mod types;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env().add_directive("info".parse()?))
        .init();

    tracing::info!("Starting ComputeCoin Matching Engine");

    let addr: SocketAddr = "[::]:50051".parse()?;
    let listener = TcpListener::bind(addr).await?;
    tracing::info!("gRPC server listening on {}", addr);

    // The full gRPC service will be registered once proto-generated code is integrated.
    // For now, accept connections to verify the server starts correctly.
    tokio::select! {
        _ = accept_loop(listener) => {}
        _ = tokio::signal::ctrl_c() => {
            tracing::info!("Received shutdown signal, gracefully stopping...");
        }
    }

    tracing::info!("Matching engine stopped.");
    Ok(())
}

async fn accept_loop(listener: TcpListener) {
    loop {
        match listener.accept().await {
            Ok((_stream, addr)) => {
                tracing::debug!("Connection from {}", addr);
            }
            Err(e) => {
                tracing::error!("Accept error: {}", e);
            }
        }
    }
}
