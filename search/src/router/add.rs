use warp::{reply::{WithStatus, Json}, http::StatusCode};
use anyhow::Result;
use crate::model;

/// This route allows to create a new document
pub async fn add(body: model::User, authorization: String) -> Result<WithStatus<Json>> {
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

    crate::database::add_document(body).await?;

    Ok(warp::reply::with_status(warp::reply::json(&model::Error {
        error: false,
        message: "Indexed".to_string(),
    }), StatusCode::CREATED))
}