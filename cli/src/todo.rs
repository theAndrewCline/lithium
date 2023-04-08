use surrealdb::{Datastore, Error, Session};

pub struct TodoStore {
    db: Datastore,
    session: Session,
}

#[derive(PartialEq, Debug, Clone)]
pub struct Todo {
    title: String,
    status: String,
}

#[derive(PartialEq)]
pub struct CreateTodoInput {
    title: String,
    status: String,
}

impl TodoStore {
    pub fn new(db: Datastore) -> Self {
        let session = Session::for_kv();

        Self { db, session }
    }

    async fn create(self: &Self, input: &CreateTodoInput) -> Result<Todo, Error> {
        let todo = Todo {
            title: input.title.clone(),
            status: input.status.clone(),
        };

        let response = self
            .db
            .execute(
                "",
                &self.session,
                vec![input.title.clone(), input.status.clone()],
                false,
            )
            .await?;

        Ok(todo)
    }
}

#[cfg(test)]
mod tests {
    use surrealdb::Response;

    use super::*;

    #[tokio::test]
    async fn create_todo() {
        let db = Datastore::new("memory").await.unwrap();

        let ses = Session::for_kv();

        let store = TodoStore::new(db);

        let create_todo_input = CreateTodoInput {
            title: String::from("a normal todo"),
            status: String::from("incomplete"),
        };

        let result = store.create(&create_todo_input).await.unwrap();

        let expected_todo = Todo {
            title: create_todo_input.title.clone(),
            status: create_todo_input.status.clone(),
        };

        assert_eq!(result, expected_todo.clone());

        let db_result = store
            .db
            .execute("select * from 'todos';", &ses, Option::None, false)
            .await
            .unwrap();
    }
}
