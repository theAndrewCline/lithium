use clap::{Args, Parser, Subcommand};

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

#[derive(Debug, Parser)]
#[command(author, version, about, long_about = None)]
pub struct Program {
    /// List todos
    #[clap(subcommand)]
    action: ActionType,
}
