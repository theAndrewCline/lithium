use std::net::SocketAddr;

use axum::{http::StatusCode, routing::get, Json, Router};
use serde::{Deserialize, Serialize};

use tracing;
use tracing_subscriber;

fn make_app() -> Router {
    Router::new().route("/", get(list_todos))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let addr = SocketAddr::from(([127, 0, 0, 1], 4000));

    let app = make_app();

    tracing::debug!("listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct Todo {
    id: String,
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct TodoResponse {
    todos: Vec<Todo>,
}

async fn list_todos() -> (StatusCode, Json<TodoResponse>) {
    (StatusCode::OK, Json(TodoResponse { todos: vec![] }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::http::StatusCode;
    use axum_test_helper::TestClient;

    #[tokio::test]
    async fn get_todos_empty() {
        let app = make_app();

        let client = TestClient::new(app);
        let res = client.get("/").send().await;

        assert_eq!(res.status(), StatusCode::OK);
        assert_eq!(
            res.json::<TodoResponse>().await,
            TodoResponse { todos: vec![] }
        );
    }
}
