import Config

config :notification,
  port: 8891,
  jwt: System.get_env("RSA_PUBLIC_KEY"),
  nats: %{host: ~c"nats", port: 4222}
