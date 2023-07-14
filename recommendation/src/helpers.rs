use jsonwebtoken::{decode, Algorithm, Validation, DecodingKey, TokenData};
use chrono::{Duration as ChronoDuration, Utc, Timelike};
use std::time::Duration;
use anyhow::Result;

/// Decode a JWT token and check if it is valid
pub fn get_jwt(token: String) -> Result<TokenData<crate::model::Claims>> {
    let public_key = DecodingKey::from_rsa_pem(dotenv::var("RSA_PUBLIC_KEY").expect("Missing env `RSA_PUBLIC_KEY`").as_bytes())
        .expect("Failed to load public key");

    Ok(decode::<crate::model::Claims>(&token, &public_key, &Validation::new(Algorithm::RS256))?)
}

/// remove_duplicates allows to delete every
/// duplicated ID in the vector
pub fn remove_duplicates(vec: &mut Vec<String>) {
    let mut unique_ids: Vec<String> = Vec::new();
    let mut encountered_ids: Vec<String> = Vec::new();

    for string in vec.iter() {
        if !encountered_ids.contains(string) {
            unique_ids.push(string.clone());
            encountered_ids.push(string.clone());
        }
    }

    *vec = unique_ids;
}

/// hourly_cron start a function that works every hour
pub async fn hourly_cron(graph: std::sync::Arc<neo4rs::Graph>) {
    tokio::task::spawn(async move {
        loop {
            let now = Utc::now();
            let time = (now.naive_utc().date().and_hms_opt(now.hour(), 0, 0).unwrap() + ChronoDuration::hours(1)).timestamp()-now.timestamp();
            std::thread::sleep(Duration::from_secs(time as u64));

            println!("Starting PageRank and Community Detection...");

            match crate::database::page_rank(graph.clone()).await {
                Ok(_) => {},
                Err(_) => eprintln!("PageRank did not work as expected")
            }

            match crate::database::community_detection(graph.clone()).await {
                Ok(_) => {},
                Err(_) => eprintln!("Community Detection did not work as expected")
            }
        }
    });
}
