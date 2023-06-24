use warp::{reply::{WithStatus, Json}, http::StatusCode};
use crate::database;
use anyhow::Result;

/// This route allows to create a new document
pub async fn users(authorization: String) -> Result<WithStatus<Json>> {
    // Check if token is valid
    if authorization != dotenv::var("GLOBAL_AUTH")? {
        return Ok(warp::reply::with_status(warp::reply::json(
            &crate::model::Error{
                error: true,
                message: "Invalid token".to_string(),
            }
        ),
        StatusCode::UNAUTHORIZED))
    }

    Ok(warp::reply::with_status(
        warp::reply::json(
            &database::get_all()
                .await?
                .hits
                .into_iter()
                .map(|u| u.result.vanity)
                .collect::<Vec<String>>()
        ), StatusCode::OK
    ))
}
