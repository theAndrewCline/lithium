use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, PartialEq, Debug)]
pub struct Todo {
    pub id: String,
    pub text: String,
}
