pub mod db_helpers;

use cuid::cuid2;
use db_helpers::{db_response_to_todo, DbResult, TodoDatabaseResponse};
use serde::{Deserialize, Serialize};
use surrealdb::engine::any::Any;
use surrealdb::Surreal;

pub static DB: Surreal<Any> = Surreal::init();

pub async fn list_todos() -> DbResult<Vec<Todo>> {
    let result: DbResult<Vec<TodoDatabaseResponse>> = DB.select("todo").await;

    let todos: DbResult<Vec<Todo>> =
        result.map(|ts| ts.iter().map(|t| db_response_to_todo(t)).collect());

    return todos;
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CreateTodoInput {
    pub text: String,
}

pub async fn create_todo(payload: CreateTodoInput) -> DbResult<TodoDatabaseResponse> {
    let result: DbResult<TodoDatabaseResponse> =
        DB.create(("todo", cuid2())).content(payload).await;

    return result;
}

pub async fn update_todo(payload: Todo) -> DbResult<TodoDatabaseResponse> {
    let todo_id = payload.id.clone();

    let result: DbResult<TodoDatabaseResponse> =
        DB.update(("todo", todo_id)).content(payload).await;

    return result;
}

pub async fn delete_todo(payload: Todo) -> DbResult<TodoDatabaseResponse> {
    let result: DbResult<TodoDatabaseResponse> = DB.delete(("todo", payload.id)).await;

    return result;
}

#[derive(Serialize, Deserialize, PartialEq, Debug)]
pub struct Todo {
    pub id: String,
    pub text: String,
}

#[cfg(test)]
mod tests {}
