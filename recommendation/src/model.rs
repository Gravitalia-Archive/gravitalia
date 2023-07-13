use serde::{Serialize, Deserialize};

/// Error defines how message with error should be
#[derive(Serialize)]
pub struct Error {
    pub error: bool,
    pub message: String
}

// Post defines how a post should be sent
#[derive(Serialize)]
pub struct Post {
	pub id: String,
	pub description: String,
	pub author: String,
    pub hash: Vec<String>,
    pub like: u32
}

// Claims defines JWT struct
#[allow(dead_code)]
#[derive(Debug, Deserialize)]
pub struct Claims {
    pub sub: String,
    pub scope: Vec<String>,
    pub exp: u64,
    iss: String,
    iat: u64
}