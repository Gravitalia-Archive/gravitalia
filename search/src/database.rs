use meilisearch_sdk::search::SearchResults;
use meilisearch_sdk::indexes::Index;
use meilisearch_sdk::client::*;
use crate::model::User;
use std::sync::Arc;
use anyhow::Result;

/// Init Meilisearch database and index
pub async fn init() -> Result<Index> {
    // Connect
    let client = Client::new(
        dotenv::var("MEILISEARCH_URL").unwrap_or_else(|_| "localhost:7700".to_string()),
        Some(dotenv::var("MEILISEARCH_KEY").unwrap_or_default())
    );

    // Create index if not exists
    client.create_index("gravitalia", Some("vanity")).await?;

    // Add sortable keys
    client.index("gravitalia").set_sortable_attributes(&["flags"]).await?;

    Ok(client.index("gravitalia"))
}

/// Allows to add a document into the index
pub async fn add_document(document: User, meili: Arc<Index>) -> Result<()> {
    // Add document
    meili.add_or_replace(&[document], Some("vanity")).await?;

    Ok(())
}

/// Allows to delete a document into the index
pub async fn delete_document(id: String, meili: Arc<Index>) -> Result<()> {
    // Add document
    meili.delete_document(id).await?;

    Ok(())
}

// Search into all documents
pub async fn search(query: String, limit: u8, meili: Arc<Index>) -> Result<SearchResults<User>> {
    Ok(
        meili
        .search()
        .with_query(&query)
        .with_sort(&[
            "flags:desc"
        ])
        .with_limit(limit.into())
        .execute::<User>()
        .await?
    )
}

// Get every documents in index
pub async fn get_all(meili: Arc<Index>) -> Result<SearchResults<User>> {
    Ok(
        meili
        .search()
        .with_query("*")
        .execute::<User>()
        .await?
    )
}
