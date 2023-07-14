use warp::{Filter, Reply, Rejection, http::StatusCode, reject::Reject};
use std::error::Error;

pub mod database;
pub mod helpers;
pub mod router;
pub mod model;

#[derive(Debug)]
struct UnknownError;
impl Reject for UnknownError {}

// This function receives a `Rejection` and tries to return a custom
// value, otherwise simply passes the rejection along.
async fn handle_rejection(err: Rejection) -> Result<impl Reply, std::convert::Infallible> {
    let code;
    let message: String;

    if err.is_not_found() {
        code = StatusCode::NOT_FOUND;
        message = "Not found".to_string();
    } else if let Some(e) = err.find::<warp::filters::body::BodyDeserializeError>() {
        message = match e.source() {
            Some(cause) => {
                cause.to_string()
            }
            None => "Invalid body".to_string(),
        };
        code = StatusCode::BAD_REQUEST;
    } else if err.find::<warp::reject::MethodNotAllowed>().is_some() {
        code = StatusCode::METHOD_NOT_ALLOWED;
        message = "Method not allowed".to_string();
    } else {
        code = StatusCode::INTERNAL_SERVER_ERROR;
        message = "Internal server error".to_string();
    }

    Ok(warp::reply::with_status(warp::reply::json(&model::Error {
        error: true,
        message,
    }), code))
}


#[tokio::main]
async fn main() {
    // Set env variables
    dotenv::dotenv().ok();

    // Init database
    let neo4j = database::init().await.unwrap();
    let oneo4j = neo4j.clone();

    // Create routes
    let routes = warp::path("recommendation")
    .and(warp::path("for_you_feed"))
    .and(warp::get())
    .and(warp::header("authorization"))
    .and(warp::any().map(move || neo4j.clone()))
    .and_then(|token: String, neo4j: std::sync::Arc<neo4rs::Graph>| async move {
        match router::foyou::get(token, neo4j).await {
            Ok(r) => {
                Ok(r)
            },
            Err(_) => {
                Err(warp::reject::custom(UnknownError))
            }
        }
    })
    .recover(handle_rejection);

    // Start CRON job
    tokio::task::spawn(async move {
        helpers::hourly_cron(oneo4j.clone()).await;
    });

    // Set port or use default
    let port: u16 = dotenv::var("RECOMMENDATION_PORT").unwrap_or_else(|_| "8889".to_string()).parse::<u16>().unwrap();
    println!("Server started on port {}", port);

    // Start server
    warp::serve(warp::any().and(warp::options()).map(|| "OK").or(warp::head().map(|| "OK")).or(routes))
    .run((
        [0, 0, 0, 0],
        port
    ))
    .await;
}
