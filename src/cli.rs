use clap::{Args, Parser, Subcommand};
use std::fmt::Error;

use crate::todo::TodoStore;

pub struct Todo {}

pub trait DatabaseAdapter {
    fn create(todo: Todo) -> Result<Todo, Error>;
    fn update(todo: Todo) -> Result<Todo, Error>;
    fn delete(todo: Todo) -> Result<Todo, Error>;
    fn list() -> Result<Vec<Todo>, Error>;
}

#[derive(Args, Debug)]
struct CreateInput {
    input: String,
}

#[derive(Args, Debug)]
struct CompleteInput {
    id: String,
}

#[derive(Args, Debug)]
struct DeleteInput {
    id: String,
}

#[derive(Debug, Subcommand)]
enum ActionType {
    /// List todos
    List,
    /// Create todo
    Create(CreateInput),
    /// Complete todo
    Complete(CompleteInput),
    /// Delete todo
    Delete(DeleteInput),
}

#[derive(Debug, Parser)]
#[command(author, version, about, long_about = None)]
pub struct Program {
    #[clap(subcommand)]
    action: ActionType,
}

pub fn run(store: &TodoStore) {
    let program = Program::parse();

    match program {
        _ => println!("it's working"),
    }
}
