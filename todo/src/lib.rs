pub mod db_helpers;

use cuid::cuid2;
use db_helpers::{db_response_to_todo, DbResult, TodoDatabaseResponse};
use serde::{Deserialize, Serialize};
use surrealdb::engine::any::Any;
use surrealdb::Surreal;

pub static DB: Surreal<Any> = Surreal::init();

pub async fn list_todos() -> DbResult<Vec<Todo>> {
    let result: DbResult<Vec<TodoDatabaseResponse>> = DB.select("todo").await;

    let todos: DbResult<Vec<Todo>> = result
        .map(|ts| ts.iter().map(|t| db_response_to_todo(t)).collect())
        .map(|mut todos: Vec<Todo>| {
            todos.sort_by(|a, b| a.referance.cmp(&b.referance));

            return todos;
        });

    return todos;
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CreateTodoPayload {
    pub text: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CreateTodoInput {
    pub text: String,
    pub referance: u32,
}

#[derive(Serialize, Deserialize, Debug)]
struct Referance {
    referance: u32,
}

pub async fn next_referance() -> u32 {
    let result: DbResult<Option<Referance>> = DB.select(("referance", "static")).await;

    let next_referance = result.expect("to have result");

    match next_referance {
        Some(r) => {
            let next_ref: DbResult<Referance> = DB
                .update(("referance", "static"))
                .content(Referance {
                    referance: r.referance + 1,
                })
                .await;

            next_ref.expect("next ref should be succesfully created");

            return r.referance;
        }
        None => {
            let next_ref: DbResult<Referance> = DB
                .update(("referance", "static"))
                .content(Referance { referance: 2 })
                .await;

            next_ref.expect("next ref should be succesfully created");

            return 1;
        }
    }
}

pub async fn create_todo(payload: CreateTodoPayload) -> DbResult<TodoDatabaseResponse> {
    let result: DbResult<TodoDatabaseResponse> = DB
        .create(("todo", cuid2()))
        .content(CreateTodoInput {
            text: payload.text,
            referance: next_referance().await,
        })
        .await;

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
    pub referance: u32,
}

#[cfg(test)]
mod tests {}
