use serde::{Deserialize, Serialize};
use serde_json;
use std::env;

#[derive(Serialize, Deserialize)]
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

async fn create_todo(input: &str) {
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

async fn complete_todo(todo: &str) {
    println!("Completed: \"{}\"", todo)
}

async fn delete_todo(todo: &str) {
    println!("Deleted: \"{}\"", todo)
}

enum Command {
    LIST,
    CREATE,
    COMPLETE,
    DELETE,
    HELP,
}

fn parse_command(args: &[String]) -> Command {
    use Command::*;
    let command = &args[1] as &str;

    match command {
        "list" => LIST,
        "create" => CREATE,
        "complete" => COMPLETE,
        "delete" => DELETE,
        _ => HELP,
    }
}

#[tokio::main]
async fn main() {
    use Command::*;
    let args: Vec<String> = env::args().collect();
    let command: Command = parse_command(&args);

    match command {
        LIST => list_todos().await,
        CREATE => create_todo(&args[2]).await,
        COMPLETE => complete_todo(&args[2]).await,
        DELETE => delete_todo(&args[2]).await,
        HELP => println!("Please provide a command"),
    }
}
