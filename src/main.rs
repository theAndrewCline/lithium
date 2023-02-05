mod cli;
mod todo;

use serde::{Deserialize, Serialize};
use serde_json;

use surrealdb::{Datastore, Error};
use todo::TodoStore;

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

#[tokio::main]
async fn main() -> Result<(), Error> {
    let store = TodoStore::new(Datastore::new("memory").await?);

    cli::run(&store);

    Ok(())
}
