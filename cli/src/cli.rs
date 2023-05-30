use std::io::{StdoutLock, Write};

use clap::{Args, Parser, Subcommand};

use todo::{create_todo, delete_todo_by_ref, list_todos, CreateTodoPayload};

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
    referance: String,
}

#[derive(Debug, Subcommand)]
enum ActionType {
    /// List todos
    List,
    /// Create todo
    Create(CreateInput),
    /// Complete todo
    // Complete(CompleteInput),
    /// Delete todo
    Delete(DeleteInput),
}

#[derive(Debug, Parser)]
#[command(author, version, about, long_about = None)]
pub struct Program {
    #[clap(subcommand)]
    action: ActionType,
}

pub async fn run(handle: &mut StdoutLock<'static>) {
    let program = Program::parse();

    match program.action {
        ActionType::List => {
            let todos = list_todos().await.expect("fetch todos");

            for todo in todos {
                writeln!(handle, "{}: {}", todo.referance, todo.text)
                    .expect("writing to io to work");
            }
        }

        ActionType::Create(input) => {
            create_todo(CreateTodoPayload {
                text: input.input.clone(),
            })
            .await
            .expect("todo to be created");

            writeln!(handle, "Created todo! {}", input.input).expect("writing should work");
        }

        ActionType::Delete(input) => match delete_todo_by_ref(input.referance).await {
            Ok(_) => writeln!(handle, "Delete successful!").expect("writing should work"),
            Err(_) => writeln!(handle, "Delete failed").expect("writing should work"),
        },
    }
}
