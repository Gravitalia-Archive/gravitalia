defmodule Notification.Router do
  use Plug.Router

  plug(:match)
  plug(:dispatch)
  plug(Notification.SSE)

  get "/sse" do
    conn |> Notification.SSE.call([]) |> send_resp(200, "")
  end

  match _ do
    send_resp(conn, 404, "404")
  end
end
