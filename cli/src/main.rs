mod cli;
mod todo;

use surrealdb::{Datastore, Error};
use todo::TodoStore;

#[tokio::main]
async fn main() -> Result<(), Error> {
    let store = TodoStore::new(Datastore::new("memory").await?);

    cli::run(&store);

    Ok(())
}
