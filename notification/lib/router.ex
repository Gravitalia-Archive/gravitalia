defmodule Notification.Router do
  @moduledoc """
  The router provides access to the right handler
  """

  use Plug.Router

  plug(:match)
  plug(:dispatch)
  plug(Notification.SSE)

  get "/sse" do
    conn |> Notification.SSE.call([]) |> send_resp(200, "")
  end

  match _ do
    conn =
      conn
      |> put_resp_header("Access-Control-Allow-Origin", "https://www.gravitalia.com")
      |> put_resp_header("Access-Control-Allow-Credentials", "true")
      |> put_resp_header("Access-Control-Allow-Methods", "GET")
      |> put_resp_header("Access-Control-Allow-Headers", "Content-Type, Authorization, Cache-Control")

    send_resp(conn, 200, "OK")
  end
end
