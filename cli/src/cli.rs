use clap::{Args, Parser, Subcommand};

use todo::list_todos;

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

pub async fn run() {
    let program = Program::parse();

    match program.action {
        ActionType::List => {
            let todos = list_todos().await.expect("fetch todos");

            for todo in todos {
                println!("{}", todo.text)
            }
        }
        _ => println!("it's working"),
    }
}
