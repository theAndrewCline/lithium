use clap::{Args, Parser, Subcommand};
use serde::{Deserialize, Serialize};
use serde_json;

#[derive(Serialize, Deserialize, Debug)]
struct Todo {
    id: uuid::Uuid,
    title: String,
    completed: bool,
}

async fn list_todos() {
    let mut files = tokio::fs::read_dir("./todos").await.unwrap();
    let mut count = 1;

    while let Some(file) = files.next_entry().await.unwrap() {
        let file_contents = tokio::fs::read(file.path()).await.unwrap();
        let todo: Todo = serde_json::from_slice(&file_contents).unwrap();

        println!("{}: {}", count, todo.title);
        count = count + 1;
    }
}

async fn create_todo(input: String) {
    let todo = Todo {
        id: uuid::Uuid::new_v4(),
        title: input.to_string(),
        completed: false,
    };

    let file_result = tokio::fs::write(
        format!("./todos/{}.json", todo.id),
        serde_json::to_string(&todo).unwrap(),
    )
    .await;

    match file_result {
        Ok(_) => println!("created todo {}", input),
        Err(e) => println!("errored because: {}", e),
    }
}

async fn complete_todo(todo: String) {
    println!("Completed: \"{}\"", todo)
}

async fn delete_todo(todo: String) {
    println!("Deleted: \"{}\"", todo)
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
    List,
    Create(CreateInput),
    Complete(CompleteInput),
    Delete(DeleteInput),
}

/// Simple program to greet a person
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct LithiumArgs {
    /// List todos
    #[clap(subcommand)]
    action: ActionType,
}

#[tokio::main]
async fn main() {
    let args = LithiumArgs::parse();

    match args.action {
        ActionType::List => list_todos().await,
        ActionType::Create(x) => create_todo(x.input).await,
        ActionType::Complete(x) => complete_todo(x.id).await,
        ActionType::Delete(x) => delete_todo(x.id).await,
    }
}
