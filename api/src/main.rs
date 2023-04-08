mod router;

use router::make_router;
use std::net::SocketAddr;

use todo::DB;

#[tokio::main]
async fn main() -> surrealdb::Result<()> {
    tracing_subscriber::fmt::init();

    let addr = SocketAddr::from(([127, 0, 0, 1], 4000));

    DB.connect("mem://").await?;
    DB.use_ns("lithium").use_db("lithium").await?;

    tracing::info!("listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(make_router())
        .await
        .unwrap();

    Ok(())
}
