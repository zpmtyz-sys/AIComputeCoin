use computecoin_matching_engine::server::proto::matching_service_server::MatchingServiceServer;
use computecoin_matching_engine::server::MatchingServiceImpl;
use tonic::transport::Server;
use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env().add_directive("info".parse()?))
        .init();

    tracing::info!("Starting ComputeCoin Matching Engine");

    let addr = "[::]:50051".parse()?;
    let service = MatchingServiceImpl::new("CU/USDT".to_string());

    tracing::info!("gRPC server listening on {}", addr);

    Server::builder()
        .add_service(MatchingServiceServer::new(service))
        .serve_with_shutdown(addr, async {
            tokio::signal::ctrl_c()
                .await
                .expect("failed to listen for ctrl_c");
            tracing::info!("Received shutdown signal, gracefully stopping...");
        })
        .await?;

    tracing::info!("Matching engine stopped.");
    Ok(())
}
