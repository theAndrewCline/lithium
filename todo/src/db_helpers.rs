use serde::{Deserialize, Serialize};
use surrealdb::sql::Thing;

use crate::Todo;

pub type DbResult<T> = Result<T, surrealdb::Error>;

#[derive(Serialize, Deserialize, PartialEq, Debug)]
pub struct TodoDatabaseResponse {
    id: Thing,
    text: String,
    referance: u32,
    complete: bool,
}

pub fn db_response_to_todo(response: &TodoDatabaseResponse) -> Todo {
    Todo {
        id: response.id.id.to_string(),
        text: response.text.clone(),
        referance: response.referance.clone(),
        complete: response.complete.clone(),
    }
}
