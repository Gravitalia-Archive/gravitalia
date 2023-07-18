use warp::{reply::{WithStatus, Json}, http::StatusCode};
use crate::database;
use anyhow::Result;

/// This route allows to get the most liked posts
pub async fn get(neo4j: std::sync::Arc<neo4rs::Graph>) -> Result<WithStatus<Json>> {
    Ok(
        warp::reply::with_status(
            warp::reply::json(
                &database::get_most_liked_posts(neo4j).await?
            ),
            StatusCode::OK
        )
    )
}
