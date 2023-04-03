use axum::{
    http::StatusCode,
    routing::{get, post, IntoMakeService},
    Json, Router,
};
use cuid::cuid2;
use serde::{Deserialize, Serialize};

use crate::{
    db_helpers::{db_response_to_todo, DbResult, TodoDatabaseResponse},
    todos::Todo,
    DB,
};

#[derive(Serialize, Deserialize, PartialEq, Debug)]
struct TodoResponse {
    todos: Vec<Todo>,
}

async fn list_todos() -> (StatusCode, Json<TodoResponse>) {
    let result: DbResult<Vec<TodoDatabaseResponse>> = DB.select("todo").await;

    let todos: DbResult<Vec<Todo>> =
        result.map(|ts| ts.iter().map(|t| db_response_to_todo(t)).collect());

    match todos {
        Ok(todos) => (StatusCode::OK, Json(TodoResponse { todos })),
        Err(err) => {
            tracing::error!("error listing todos: {}", err);

            return (StatusCode::OK, Json(TodoResponse { todos: vec![] }));
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

    let result: DbResult<Vec<TodoDatabaseResponse>> = DB.update("todo").content(payload).await;

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
