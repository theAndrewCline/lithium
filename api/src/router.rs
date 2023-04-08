use axum::{
    http::StatusCode,
    routing::{get, post, IntoMakeService},
    Json, Router,
};
use cuid::cuid2;
use serde::{Deserialize, Serialize};
use todo::Todo;

use crate::{
    db_helpers::{db_response_to_todo, DbResult, TodoDatabaseResponse},
    DB,
};

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

async fn list_todos() -> (StatusCode, Json<ApiResponse>) {
    let result: DbResult<Vec<TodoDatabaseResponse>> = DB.select("todo").await;

    let todos: DbResult<Vec<Todo>> =
        result.map(|ts| ts.iter().map(|t| db_response_to_todo(t)).collect());

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

#[derive(Serialize, Deserialize, Debug)]
struct CreateTodoInput {
    pub text: String,
}

async fn create_todo(Json(payload): Json<CreateTodoInput>) -> StatusCode {
    tracing::info!("payload: {:?}", payload);
    let result: DbResult<Option<TodoDatabaseResponse>> =
        DB.create(("todo", cuid2())).content(payload).await;

    match result {
        Ok(_) => StatusCode::CREATED,
        Err(err) => {
            tracing::error!("error creating todo: {}", err);
            return StatusCode::INTERNAL_SERVER_ERROR;
        }
    }
}

async fn update_todo(Json(payload): Json<Todo>) -> StatusCode {
    tracing::info!("payload: {:?}", payload);

    let todo_id = payload.id.clone();

    let result: DbResult<TodoDatabaseResponse> =
        DB.update(("todo", todo_id)).content(payload).await;

    match result {
        Ok(_) => StatusCode::OK,
        Err(err) => {
            tracing::error!("error updating todo: {}", err);

            StatusCode::INTERNAL_SERVER_ERROR
        }
    }
}

async fn delete_todo(Json(payload): Json<Todo>) -> StatusCode {
    tracing::info!("payload: {:?}", payload);

    let result: DbResult<Option<TodoDatabaseResponse>> = DB.delete(("todo", payload.id)).await;

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
        .route("/", get(list_todos))
        .route("/create", post(create_todo))
        .route("/update", post(update_todo))
        .route("/delete", post(delete_todo))
        .into_make_service()
}
