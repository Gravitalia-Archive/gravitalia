use warp::{reply::{WithStatus, Json}, http::StatusCode};
use anyhow::Result;
use crate::model;

/// This route allows to create a new document
pub async fn delete(body: model::User, authorization: String, meili: std::sync::Arc<meilisearch_sdk::indexes::Index>) -> Result<WithStatus<Json>> {
    // Check if token is valid
    if authorization != dotenv::var("GLOBAL_AUTH")? {
        return Ok(warp::reply::with_status(warp::reply::json(
            &model::Error{
                error: true,
                message: "Invalid token".to_string(),
            }
        ),
        StatusCode::UNAUTHORIZED))
    }

    match crate::database::delete_document(body.vanity, meili).await {
        Ok(_) => {},
        Err(e) => {
            eprintln!("Deleting error: {}", e);
        }
    }

    Ok(warp::reply::with_status(warp::reply::json(&model::Error {
        error: false,
        message: "Deleted".to_string(),
    }), StatusCode::OK))
}