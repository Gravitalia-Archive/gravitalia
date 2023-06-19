use meilisearch_sdk::search::SearchResults;
use meilisearch_sdk::indexes::Index;
use meilisearch_sdk::client::*;
use once_cell::sync::OnceCell;
use crate::model::User;
use anyhow::Result;

static INDEX: OnceCell<Index> = OnceCell::new();

/// Init Meilisearch database and index
pub async fn init() -> Result<()> {
    // Connect
    let client = Client::new(
        dotenv::var("MEILISEARCH_URL").unwrap_or_else(|_| "localhost:7700".to_string()),
        Some(dotenv::var("MEILISEARCH_URL").unwrap_or_default())
    );

    // Create index if not exists
    client.create_index("gravitalia", Some("vanity")).await?;

    // Set index
    let _ = INDEX.set(client.index("gravitalia"));

    Ok(())
}

/// Allows to add a document into the index
pub async fn add_document(document: User) -> Result<()> {
    // Add document
    INDEX.get().unwrap().add_or_replace(&[document], Some("vanity")).await?;

    Ok(())
}

/// Allows to delete a document into the index
pub async fn delete_document(id: String) -> Result<()> {
    // Add document
    INDEX.get().unwrap().delete_document(id).await?;

    Ok(())
}

// Search into all documents
pub async fn search(query: String) -> Result<SearchResults<User>> {
    Ok(
        INDEX.get().unwrap()
        .search()
        .with_query(&query)
        .with_limit(3)
        .execute::<User>()
        .await?
    )
}