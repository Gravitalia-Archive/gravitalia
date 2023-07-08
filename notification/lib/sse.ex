defmodule Notification.SSE do
  @behaviour Plug

  import Plug.Conn

  def init(opts), do: opts

  def call(conn, opts) do
    conn =
      conn
      |> put_resp_header("Access-Control-Allow-Origin", "*")
      |> put_resp_header("Access-Control-Allow-Methods", "GET")
      |> put_resp_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
      |> put_resp_header("Cache-Control", "no-cache")
      |> put_resp_header("connection", "keep-alive")
      |> put_resp_header("Content-Type", "text/event-stream; charset=utf-8")
      |> send_chunked(200)

    PubSub.subscribe(self(), "user_id")

    sse_loop(conn, self())
  end

  defp sse_loop(conn, pid) do
    receive do
      {message} ->
        chunk(conn, "id:#{UUID.uuid1()}\nevent: message\ndata: #{Jason.encode!(message)}\n\n")

        sse_loop(conn, pid)

      {:DOWN, _reference, :process, ^pid, _type} ->
        nil

      _other ->
        sse_loop(conn, pid)
    end
  end
end
