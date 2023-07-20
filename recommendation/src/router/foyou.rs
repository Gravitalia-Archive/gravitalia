use warp::{reply::{WithStatus, Json}, http::StatusCode};
use crate::database;
use anyhow::Result;
use crate::model;

const QUERY_LAST_COMMUNITY_POST: &str = "MATCH (u:User {name: $id}) WITH u MATCH (a:User {community: u.community})-[r]->(p:Post) WHERE NOT EXISTS((u)-[:VIEW]->(p)) WITH p, count(r) as connections ORDER BY connections DESC LIMIT 100 WITH p ORDER BY p.id DESC LIMIT 30 RETURN p;";
const QUERY_LAST_FOLLOWING_POST: &str = "MATCH (n:User {name: $id})-[:SUBSCRIBER]->(u:User) MATCH (u)-[:CREATE]->(p:Post) WHERE NOT EXISTS((n)-[:VIEW]->(p)) WITH p ORDER BY p.id DESC LIMIT 20 RETURN p;";
const QUERY_LAST_LIKED_POST: &str = "MATCH (u:User {name: $id})-[:LIKE]->(p:Post)-[:SHOW]->(t:Tag) WITH u, p, t ORDER BY p.id DESC LIMIT 1 WITH u, t MATCH (p:Post)-[:SHOW]->(t:Tag) WHERE NOT EXISTS((u)-[:VIEW]->(p)) WITH p ORDER BY p.id DESC LIMIT 10 RETURN p;";

/// This route finds most revelant posts to the user and then
/// send them
pub async fn get(token: String, neo4j: std::sync::Arc<neo4rs::Graph>) -> Result<WithStatus<Json>> {
    let vanity = match crate::helpers::get_jwt(token) {
        Ok(claims) => {
            claims.claims.sub
        },
        Err(e) => {
            eprintln!("{}", e);
            return Ok(warp::reply::with_status(warp::reply::json(&model::Error {
                error: true,
                message: crate::router::INVALID_TOKEN.to_string(),
            }), StatusCode::BAD_REQUEST));
        }
    };

    // Get posts with same tag as last liked post
    let tag_post = match database::last_x_post(
        neo4j.clone(),
        QUERY_LAST_LIKED_POST.to_string(),
        vanity.clone()
    ).await {
        Ok(posts) => posts,
        Err(e)=> {
            eprintln!("Cannot get latest liked posts: {}", e);

            return Ok(warp::reply::with_status(warp::reply::json(&model::Error {
                error: true,
                message: crate::router::CANNOT_GET_LATEST_LIKED_POSTS.to_string(),
            }), StatusCode::BAD_REQUEST));
        }
    };

    // Get posts from following users
    let following_post = match database::last_x_post(
        neo4j.clone(),
        QUERY_LAST_FOLLOWING_POST.to_string(),
        vanity.clone()
    ).await {
        Ok(posts) => posts,
        Err(e)=> {
            eprintln!("Cannot get latest following posts: {}", e);

            return Ok(warp::reply::with_status(warp::reply::json(&model::Error {
                error: true,
                message: crate::router::CANNOT_GET_LATEST_FOLLOWING_POSTS.to_string(),
            }), StatusCode::BAD_REQUEST));
        }
    };

    // Get posts from the user's community
    let community_post = match database::last_x_post(
        neo4j.clone(),
        QUERY_LAST_COMMUNITY_POST.to_string(),
        vanity.clone()
    ).await {
        Ok(posts) => posts,
        Err(e)=> {
            eprintln!("Cannot get latest community posts: {}", e);

            return Ok(warp::reply::with_status(warp::reply::json(&model::Error {
                error: true,
                message: crate::router::CANNOT_GET_LATEST_FOLLOWING_POSTS.to_string(),
            }), StatusCode::BAD_REQUEST));
        }
    };

    let mut posts: Vec<String> = Vec::new();
    posts.extend(tag_post);
    posts.extend(following_post);
    posts.extend(community_post);

    if posts.len() == 0 {
        return Ok(warp::reply::with_status(warp::reply::json(&posts), StatusCode::OK));
    }

    crate::helpers::remove_duplicates(&mut posts);

    Ok(
        warp::reply::with_status(
            warp::reply::json(
                &database::jaccard_index(neo4j, vanity, posts).await?
            ),
            StatusCode::OK
        )
    )
}
