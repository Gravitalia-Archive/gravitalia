<img src="https://avatars.githubusercontent.com/u/81774317?s=200&v=4" width="40" />

# Gravitalia
> Gravitalia is a social network for publishing photos between friends. 🔥<br>
> We are always trying to improve performance and safety and we rely on our best members to help us! 💪

[![Discord](https://img.shields.io/discord/843780677019500565?label=Chat&logo=discord&style=for-the-badge[Discord])](https://discord.gg/4dcEwKj2KM)

# Database
## Memgraph
> Memgraph is an in-memory graph database compatible with Neo4j

Used for posts (*photos*), users and their edges.
<img src="https://media.discordapp.net/attachments/844241319165558803/1057767661976690799/Capture_decran_2022-12-28_a_22.10.47.png" />

## Memcached
> Memcached is a key-value in-memory database

Used for cache recent count (*followers, following...*) and `states` for OAuth query.

# Security
> **This service DOESN'T store ANY sensitive data**
## JWT
- Token for maximum 7 days
- RSA Key

# Privacy
> For Gravitalia, privacy is important!

We do not sell your data, do not use it for advertising purposes.

Our recommendation service is only based on what you **liked** and your **subscriptions**, and only according to **THIS platform**!
