<img src="https://avatars.githubusercontent.com/u/81774317?s=200&v=4" width="40" />

# Gravitalia
> Gravitalia is a social network for publishing photos between friends. 🔥<br>
> We are always trying to improve performance and safety and we rely on our best members to help us! 💪

[![Discord](https://img.shields.io/discord/843780677019500565?label=Chat&logo=discord&style=for-the-badge[Discord])](https://discord.gg/4dcEwKj2KM)

## Architecture
Gravitalia is separated into two folders to facilitate the [microservices](https://en.wikipedia.org/wiki/Microservices). Here is a description of the two files:

| Name | Readme | Description |
|------------|------------|------------|
| Rest | [readme](rest/README.md) | Gravitalia's main API that allows you to delete your account, obtain data, add photos... |
| Recommendation | [readme](recommendation/README.md) | Recommendation system based on several algorithms and [Memgraph](https://memgraph.com/) |

## Use the code
To use the code efficiently, you can download the docker-compose.yml and type `docker-compose up`.

⚠️ Check if [Docker](https://www.docker.com/) is installed ⚠️
