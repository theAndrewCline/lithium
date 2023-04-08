use axum::{
    http::StatusCode,
    routing::{get, post, IntoMakeService},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use todo::{create_todo, delete_todo, list_todos, update_todo, CreateTodoInput, Todo};

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct TodoListResponse {
    todos: Vec<Todo>,
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct ErrorResponse {
    message: String,
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
enum ApiResponse {
    Todos(Vec<Todo>),
    Todo(Todo),
    Error(String),
}

async fn list_todos_route() -> (StatusCode, Json<ApiResponse>) {
    let todos = list_todos().await;

    match todos {
        Ok(todos) => (StatusCode::OK, Json(ApiResponse::Todos(todos))),
        Err(err) => {
            tracing::error!("error listing todos: {}", err);

            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ApiResponse::Error(String::from("could not get todos"))),
            );
        }
    }
}

async fn create_todo_route(Json(payload): Json<CreateTodoInput>) -> StatusCode {
    tracing::info!("payload: {:?}", payload);

    let result = create_todo(payload).await;

    match result {
        Ok(_) => StatusCode::CREATED,
        Err(err) => {
            tracing::error!("error creating todo: {}", err);
            return StatusCode::INTERNAL_SERVER_ERROR;
        }
    }
}

async fn update_todo_route(Json(payload): Json<Todo>) -> StatusCode {
    tracing::info!("payload: {:?}", payload);

    let result = update_todo(payload).await;

    match result {
        Ok(_) => StatusCode::OK,
        Err(err) => {
            tracing::error!("error updating todo: {}", err);

            StatusCode::INTERNAL_SERVER_ERROR
        }
    }
}

async fn delete_todo_route(Json(payload): Json<Todo>) -> StatusCode {
    tracing::info!("payload: {:?}", payload);

    let result = delete_todo(payload).await;

    match result {
        Ok(_) => StatusCode::OK,
        Err(err) => {
            tracing::error!("error updating todo: {}", err);

            StatusCode::INTERNAL_SERVER_ERROR
        }
    }
}

pub fn make_router() -> IntoMakeService<Router> {
    Router::new()
        .route("/", get(list_todos_route))
        .route("/create", post(create_todo_route))
        .route("/update", post(update_todo_route))
        .route("/delete", post(delete_todo_route))
        .into_make_service()
}
