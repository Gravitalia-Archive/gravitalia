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
    send_resp(conn, 200, "OK")
  end
end
