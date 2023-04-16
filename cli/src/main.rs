mod cli;
use surrealdb::Error;

use cli::run;

use todo::DB;

#[tokio::main]
async fn main() -> Result<(), Error> {
    let db_directory = format!(
        "{}{}/.lithium",
        "file://",
        std::env::var("HOME").expect("HOME directory should be defined")
    );

    DB.connect(db_directory).await?;

    DB.use_ns("lithium").use_db("lithium").await?;

    run().await;

    Ok(())
}
