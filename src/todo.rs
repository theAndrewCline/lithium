use surrealdb::{Datastore, Error, Session};

pub struct TodoStore {
    db: Datastore,
    session: Session,
}

pub struct Todo {
    title: String,
    status: String,
}

pub struct CreateTodoInput {
    title: String,
    status: String,
}

impl TodoStore {
    pub fn new(db: Datastore) -> Self {
        let session = Session::for_kv();

        Self { db, session }
    }
}
