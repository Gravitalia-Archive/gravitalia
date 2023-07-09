import Config

config :notification,
  port: 8891,
  jwt: System.get_env("RSA_PUBLIC_KEY")
