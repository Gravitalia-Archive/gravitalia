<img src="https://avatars.githubusercontent.com/u/81774317?s=200&v=4" width="40" />

# Gravitalia
> Gravitalia is a social network for publishing photos between friends. 🔥<br>
> We are always trying to improve performance and safety and we rely on our best members to help us! 💪

[![Discord](https://img.shields.io/discord/843780677019500565?label=Chat&logo=discord&style=for-the-badge[Discord])](https://discord.gg/4dcEwKj2KM)

## Notification
> In-app (and soon push app) notification service to manage notifications of likes, new comments and subscription requests.

## Implementation
#### > To ensure a scalable, reliable and high-performance system, we chose to use Elixir.
The lightness and independence of the Elixir processes means we can manage several thousand requests in record time, and also withstand breakdowns. Indeed, if one of these processes crashes, thanks to the "one for one" strategy, we can continue to manage the other users, and recreate the process that has failed.

#### > To guarantee real-time notifications, we chose SSE.
Thanks to the ease with which Server-Sent Events (**SSE**) can be integrated, we were able to deploy our solution quickly and reliably. The one-way nature of the information transmitted reinforced the excellence of our choice and approach with SSE.


## Architecture
<img src="https://raw.githubusercontent.com/Gravitalia/.github/main/gravitalia/Notification_Architecture.png" />
