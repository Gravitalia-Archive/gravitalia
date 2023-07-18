use neo4rs::{Graph, Node, query as neo_query};
use crate::model::Post;
use anyhow::Result;
use std::sync::Arc;

/// Init database connection and send it
pub async fn init() -> Result<Arc<Graph>> {
    Ok(
        Arc::new(
            Graph::new(
                dotenv::var("GRAPH_URL").unwrap_or_else(|_| "bolt://127.0.0.1:7687".to_string()),
                dotenv::var("GRAPH_USERNAME").unwrap_or_default(),
                dotenv::var("GRAPH_PASSWORD").unwrap_or_default()
            )
            .await?
        )
    )
}

/// Starts a new calculation of PageRank in the database
pub async fn page_rank(graph: Arc<Graph>) -> Result<()> {
    graph.execute(
        neo_query("MATCH p=(n:User)-[r]->(m:User) WHERE type(r) <> 'Block' WITH project(p) as graph CALL pagerank_online.update(graph) YIELD node, rank SET node.rank = rank;")
    ).await?;

    Ok(())
}

/// Starts a new calculation of community detection in the database
pub async fn community_detection(graph: Arc<Graph>) -> Result<()> {
    graph.execute(
        neo_query("MATCH p=(n:User)-[r]->(m) WHERE type(r) <> 'Block' AND type(r) <> 'View' WITH project(p) as graph CALL community_detection_online.update(graph) YIELD node, community_id WITH node, community_id WHERE labels(node) = ['User'] SET node.community = community_id;")
    ).await?;

    Ok(())
}

/// last_x_post lets you decide what you want
/// to get from database by setting your own query.
/// You'll need to set an output of p (as Post)
pub async fn last_x_post(graph: Arc<Graph>, query: String, id: String) -> Result<Vec<std::string::String>> {
    let ids = tokio::spawn(async move {
        let mut result = graph.execute(
            neo_query(query.as_str())
            .param("id", id)
        ).await.unwrap();

        let mut id_list: Vec<String> = Vec::new();

        while let Ok(Some(row)) = result.next().await {
            match row.get::<Node>("p") {
                Some(p) => {
                    match p.get::<String>("id") {
                        Some(id) => {
                            id_list.push(
                                id
                            )
                        },
                        None => {}
                    }
                },
                None => {}
            }
        }

        id_list
    });

    Ok(ids.await?)
}

/// jaccard_index ranks every id in idList with the
/// Jaccard similarity algorithm
pub async fn jaccard_index(graph: Arc<Graph>, id: String, ids: Vec<String>) -> Result<Vec<Post>> {
    let ids = tokio::spawn(async move {
        let mut result = graph.execute(
            neo_query("MATCH (u:User {name: $id})-[:Like]->(p:Post) WITH u, p LIMIT 10 MATCH (l:Post) WHERE l.id IN $list WITH l, p ORDER BY p.id DESC WITH collect(l) as posts, collect(p) as likedPosts CALL node_similarity.jaccard_pairwise(posts, likedPosts) YIELD node1, node2, similarity WITH node1, similarity ORDER BY similarity DESC LIMIT 15 OPTIONAL MATCH (a:User)-[:Like]->(node1) WITH node1, count(DISTINCT a) as numLikes MATCH (creator:User)-[:Create]-(node1) WITH node1, numLikes, creator OPTIONAL MATCH (:User {name: $id})-[r:Like]-(node1) RETURN node1 as p, numLikes, creator, CASE WHEN r IS NULL THEN false ELSE true END;")
            .param("id", id)
            .param("list", ids)
        ).await.unwrap();

        let mut post_list: Vec<Post> = Vec::new();

        while let Ok(Some(row)) = result.next().await {
            let node = row.get::<Node>("p").unwrap();

            post_list.push(
                Post {
                    id: node.get::<String>("id").unwrap(),
                    description: node.get::<String>("description").unwrap(),
                    author: row.get::<Node>("creator").unwrap().get::<String>("name").unwrap(),
                    hash: node.get::<Vec<String>>("hash").unwrap(),
                    like: row.get::<i64>("numLikes").unwrap() as u32
                }
            )
        }

        post_list
    });

    Ok(ids.await?)
}

/// get_most_liked_posts returns the twenty most liked
/// posts on the database
pub async fn get_most_liked_posts(graph: Arc<Graph>) -> Result<Vec<Post>> {
    let ids = tokio::spawn(async move {
        let mut result = graph.execute(
            neo_query("MATCH (u:User)-[:Create]->(p:Post)<-[r:Like]-(:User) WITH p, count(DISTINCT r) as numLikes, u.name AS author ORDER BY numLikes DESC LIMIT 20 RETURN p, numLikes, author;")
        ).await.unwrap();

        let mut post_list: Vec<Post> = Vec::new();

        while let Ok(Some(row)) = result.next().await {
            let node = row.get::<Node>("p").unwrap();

            post_list.push(
                Post {
                    id: node.get::<String>("id").unwrap(),
                    description: node.get::<String>("description").unwrap(),
                    author: row.get::<String>("author").unwrap(),
                    hash: node.get::<Vec<String>>("hash").unwrap(),
                    like: row.get::<i64>("numLikes").unwrap() as u32
                }
            )
        }

        post_list
    });

    Ok(ids.await?)
}