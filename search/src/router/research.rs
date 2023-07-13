use warp::{reply::{WithStatus, Json}, http::StatusCode};
use crate::database;
use anyhow::Result;
use crate::model;

/// This route allows to search in all documents
pub async fn research(query: model::QuerySearch, meili: std::sync::Arc<meilisearch_sdk::indexes::Index>) -> Result<WithStatus<Json>> {
    if query.q == "*" {
        return Ok(warp::reply::with_status(warp::reply::json(&model::Error {
            error: true,
            message: "Cannot search '*'".to_string(),
        }), StatusCode::BAD_REQUEST));
    }

    Ok(warp::reply::with_status(
        warp::reply::json(
            &database::search(query.q, query.limit.unwrap_or(3).max(20), meili)
                .await?
                .hits
                .into_iter()
                .map(|u| u.result.vanity)
                .collect::<Vec<String>>()
        ), StatusCode::OK
    ))
}