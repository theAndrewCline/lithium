use axum::{
    http::StatusCode,
    routing::{get, post, IntoMakeService},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use todo::{create_todo, delete_todo_by_id, list_todos, update_todo, CreateTodoPayload, Todo};

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

async fn list_todos_route(
) -> Result<(StatusCode, Json<ApiResponse>), (StatusCode, Json<ApiResponse>)> {
    list_todos()
        .await
        .map(|todos| (StatusCode::OK, Json(ApiResponse::Todos(todos))))
        .map_err(|err| {
            tracing::error!("error listing todos: {:?}", err);

            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ApiResponse::Error(String::from("could not get todos"))),
            );
        })
}

async fn create_todo_route(
    Json(payload): Json<CreateTodoPayload>,
) -> Result<StatusCode, StatusCode> {
    create_todo(payload)
        .await
        .map(|_| StatusCode::CREATED)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

async fn update_todo_route(Json(payload): Json<Todo>) -> Result<StatusCode, StatusCode> {
    update_todo(payload)
        .await
        .map(|_| StatusCode::OK)
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct TodoId {
    id: String,
}

async fn delete_todo_route(Json(payload): Json<TodoId>) -> Result<StatusCode, StatusCode> {
    delete_todo_by_id(payload.id)
        .await
        .map(|_| StatusCode::OK)
        .map_err(|err| {
            tracing::error!("error updating todo: {}", err);

            StatusCode::INTERNAL_SERVER_ERROR
        })
}

pub fn make_router() -> IntoMakeService<Router> {
    Router::new()
        .route("/", get(list_todos_route))
        .route("/create", post(create_todo_route))
        .route("/update", post(update_todo_route))
        .route("/delete", post(delete_todo_route))
        .into_make_service()
}
