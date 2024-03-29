pub mod db_helpers;

use cuid::cuid2;
use db_helpers::{db_response_to_todo, TodoDatabaseResponse};
use serde::{Deserialize, Serialize};
use surrealdb::engine::any::Any;
use surrealdb::Surreal;

#[derive(Serialize, Deserialize, PartialEq, Debug)]
pub struct Todo {
    pub id: String,
    pub text: String,
    pub referance: u32,
    pub complete: bool,
}

pub static DB: Surreal<Any> = Surreal::init();

pub async fn list_todos() -> Result<Vec<Todo>, LithiumError> {
    Ok(Into::<Vec<TodoDatabaseResponse>>::into(
        DB.select("todo")
            .await
            .map_err(|err| LithiumError::Db(err.to_string()))?,
    )
    .iter()
    .map(|t| db_response_to_todo(t))
    .collect::<Vec<Todo>>())
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct CreateTodoPayload {
    pub text: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct CreateTodoInput {
    pub text: String,
    pub referance: u32,
    pub complete: bool,
}

#[derive(Serialize, Deserialize, Debug)]
struct Referance {
    referance: u32,
}

#[derive(Debug)]
pub enum LithiumError {
    Db(String),
    NotFound,
}

pub async fn next_referance() -> Result<u32, LithiumError> {
    let next_ref: Result<Referance, surrealdb::Error> = DB.select(("referance", "static")).await;

    match next_ref {
        Ok(r) => {
            Into::<Result<Referance, LithiumError>>::into(
                DB.update(("referance", "static"))
                    .content(Referance {
                        referance: r.referance + 1,
                    })
                    .await
                    .map_err(|err| LithiumError::Db(err.to_string())),
            )?;

            return Ok(r.referance);
        }
        Err(_) => {
            Into::<Result<Referance, LithiumError>>::into(
                DB.update(("referance", "static"))
                    .content(Referance { referance: 2 })
                    .await
                    .map_err(|err| LithiumError::Db(err.to_string())),
            )?;

            return Ok(1);
        }
    }
}

pub async fn create_todo(payload: CreateTodoPayload) -> Result<TodoDatabaseResponse, LithiumError> {
    let referance = next_referance().await?;

    let result: Result<TodoDatabaseResponse, LithiumError> = DB
        .create(("todo", cuid2()))
        .content(CreateTodoInput {
            text: payload.text.clone(),
            referance,
            complete: false,
        })
        .await
        .map_err(|err| LithiumError::Db(err.to_string()));

    return result;
}

pub async fn update_todo(payload: Todo) -> Result<TodoDatabaseResponse, surrealdb::Error> {
    let todo_id = &payload.id;

    let result: Result<TodoDatabaseResponse, surrealdb::Error> =
        DB.update(("todo", todo_id)).content(payload).await;

    return result;
}

pub async fn delete_todo_by_id(id: String) -> Result<TodoDatabaseResponse, surrealdb::Error> {
    DB.delete(("todo", id)).await
}

pub async fn delete_todo_by_ref(ref_str: String) -> Result<(), LithiumError> {
    let select_result: Result<Option<TodoDatabaseResponse>, surrealdb::Error> = DB
        .query("SELECT * FROM todo WHERE referance = $ref")
        .bind(("ref", ref_str))
        .await
        .map(|mut r| r.take(0))
        .map_err(|err| LithiumError::Db(err.to_string()))?;

    select_result.map_err(|_| LithiumError::NotFound)?;

    Ok(())
}

pub async fn complete_todo_by_ref(_ref_str: String) -> Result<(), LithiumError> {
    Ok(())
}
