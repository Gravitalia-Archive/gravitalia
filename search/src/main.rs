use warp::{Filter, Reply, Rejection, http::StatusCode, reject::Reject};
use std::error::Error;

pub mod model;
pub mod router;
pub mod database;

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
    // Init database
    database::init().await.unwrap();

    // Create routes
    let routes = warp::path("add")
                    .and(warp::post())
                    .and(warp::body::json())
                    .and(warp::header("authorization"))
                    .and_then(|body: model::User, token: String| async move {
                        match router::add::add(body, token).await {
                            Ok(r) => {
                                Ok(r)
                            },
                            Err(_) => {
                                Err(warp::reject::custom(UnknownError))
                            }
                        }
                    })
                .recover(handle_rejection);

    // Set port or use default
    let port: u16 = dotenv::var("PORT").unwrap_or_else(|_| "8890".to_string()).parse::<u16>().unwrap();
    println!("Server started on port {}", port);

    // Start server
    warp::serve(warp::any().and(warp::options()).map(|| "OK").or(warp::head().map(|| "OK")).or(routes))
    .run((
        [0, 0, 0, 0],
        port
    ))
    .await;
}