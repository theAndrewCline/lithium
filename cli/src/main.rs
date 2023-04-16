mod cli;

use cli::run;
use std::io::{self, StdoutLock, Write};
use todo::DB;

#[tokio::main]
async fn main() {
    let db_directory = format!(
        "{}{}/.lithium",
        "file://",
        std::env::var("HOME").expect("HOME directory should be defined")
    );

    let stdout = io::stdout(); // get the global stdout entity
    let mut handle: StdoutLock<'static> = stdout.lock(); // acquire a lock on it

    DB.connect(db_directory).await.expect("database connection");

    DB.use_ns("lithium")
        .use_db("lithium")
        .await
        .expect("namespace declaration");

    run(&mut handle).await;

    handle.flush().expect("write to stdout");
}
