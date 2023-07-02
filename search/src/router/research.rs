use warp::{reply::{WithStatus, Json}, http::StatusCode};
use crate::database;
use anyhow::Result;
use crate::model;

/// This route allows to search in all documents
pub async fn research(query: model::QuerySearch) -> Result<WithStatus<Json>> {
    if query.q == "*" {
        return Ok(warp::reply::with_status(warp::reply::json(&model::Error {
            error: true,
            message: "Cannot search '*'".to_string(),
        }), StatusCode::BAD_REQUEST));
    }

    match database::search(query.q, query.limit.unwrap_or(3).max(20)).await {
        Ok(d) => {
            Ok(warp::reply::with_status(
                warp::reply::json(
                    d.hits
                        .into_iter()
                        .map(|u| u.result.vanity)
                        .collect::<Vec<String>>()
                ), StatusCode::OK
            ))
        },
        Err(e) => {
            eprintln!(e);
            Err(e)
        }
    }
}
