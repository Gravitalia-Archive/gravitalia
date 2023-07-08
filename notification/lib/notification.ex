defmodule Notification do
  use Application
  require Logger

  def start(_type, _args) do
    children = [
      Plug.Cowboy.child_spec(
        scheme: :http,
        plug: Notification.Router,
        options: [
          port: 8891
        ],
        protocol_options: [idle_timeout: :infinity]
      ),
      Registry.child_spec(
        keys: :duplicate,
        name: Registry.Notification
      ),
      {PubSub, []}
    ]

    Logger.info("Server started at http://localhost:8891")

    opts = [strategy: :one_for_one, name: Notification.Application]
    Supervisor.start_link(children, opts)
  end
end
