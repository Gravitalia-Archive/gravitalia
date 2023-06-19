use serde::{Serialize, Deserialize};

/// Error struct defines how message with
/// error must be
#[derive(Serialize)]
pub struct Error {
    pub error: bool,
    pub message: String,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct User {
    pub vanity: String
}