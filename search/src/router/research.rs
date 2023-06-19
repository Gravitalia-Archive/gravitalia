use warp::{reply::{WithStatus, Json}, http::StatusCode};
use crate::database;
use anyhow::Result;

/// This route allows to search in all documents
pub async fn research(query: String) -> Result<WithStatus<Json>> {
    if query == "*" {
        return Ok(warp::reply::with_status(warp::reply::json(&crate::model::Error {
            error: true,
            message: "Cannot search '*'".to_string(),
        }), StatusCode::BAD_REQUEST));
    }

    Ok(warp::reply::with_status(
        warp::reply::json(
            &database::search(query)
                .await?
                .hits
                .into_iter()
                .map(|u| u.result.vanity)
                .collect::<String>()
        ), StatusCode::OK
    ))
}