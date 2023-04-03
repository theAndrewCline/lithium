use crate::todos::Todo;

use serde::{Deserialize, Serialize};
use surrealdb::sql::Thing;

pub type DbResult<T> = Result<T, surrealdb::Error>;

#[derive(Serialize, Deserialize, PartialEq, Debug)]
pub struct TodoDatabaseResponse {
    id: Thing,
    text: String,
}

pub fn db_response_to_todo(response: &TodoDatabaseResponse) -> Todo {
    Todo {
        id: response.id.to_string(),
        text: response.text.clone(),
    }
}
