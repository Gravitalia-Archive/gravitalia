defmodule Notification do
  import Logger
  use Application

  def start(_type, _args) do
    children = [
      Plug.Cowboy.child_spec(
        scheme: :http,
        plug: Notification.Router,
        options: [
          port: 4000
        ],
        protocol_options: [idle_timeout: :infinity]
      ),
      Registry.child_spec(
        keys: :duplicate,
        name: Registry.Notification
      ),
      {PubSub, []}
    ]

    Logger.info("Server started at http://localhost:4000")

    opts = [strategy: :one_for_one, name: Notification.Application]
    Supervisor.start_link(children, opts)
  end
end
