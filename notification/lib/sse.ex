defmodule Notification.SSE do
  @moduledoc """
  SSE is an implementation of Server Sent Events technology used to send notifications
  """
  @behaviour Plug

  import Plug.Conn

  def init(opts), do: opts

  defp start_subscription(user_id) do
    spawn(fn -> Notification.Nats.start_subscription(user_id) end)
  end

  defp heartbeat(user_id) do
    spawn(fn -> heartbeat_loop(user_id) end)
  end

  defp heartbeat_loop(user_id) do
    :timer.sleep(50_000)
    PubSub.publish(user_id, {Jason.encode!(%{type: "PING"})})
    heartbeat_loop(user_id)
  end

  def call(conn, _opts) do
    conn =
      conn
      |> put_resp_header("Access-Control-Allow-Origin", "https://www.gravitalia.com")
      |> put_resp_header("Access-Control-Allow-Credentials", "true")
      |> put_resp_header("Access-Control-Allow-Methods", "GET")
      |> put_resp_header("Access-Control-Allow-Headers", "Content-Type, Authorization, Cache-Control")
      |> put_resp_header("Access-Control-Max-Age", "3600")
      |> put_resp_header("Cache-Control", "no-cache")
      |> put_resp_header("connection", "keep-alive")

    case get_token(conn) do
      nil ->
        unauthorized(conn)

      token ->
        case Notification.Helpers.check_token(token) do
          nil ->
            unauthorized(conn)

          user_id ->
            conn =
              conn
              |> put_resp_header("Content-Type", "text/event-stream; charset=utf-8")
              |> send_chunked(200)

            PubSub.subscribe(self(), user_id)
            start_subscription(user_id)
            heartbeat(user_id)

            Process.send_after(self(), :stop_heartbeat, 10 * 60 * 1000)

            sse_loop(conn, self())
        end
    end
  end

  defp unauthorized(conn) do
    conn =
      conn
      |> put_resp_header("Content-Type", "application/json")
      |> send_chunked(401)
      |> chunk(Jason.encode!(%{error: true, message: "Invalid token"}))

    conn
  end

  defp sse_loop(conn, pid) do
    receive do
      {message} ->
        chunk(conn, format_sse_message(message))
        sse_loop(conn, pid)

      {:DOWN, _reference, :process, ^pid, _type} ->
        nil

      :stop_heartbeat ->
        Process.exit(pid, :normal)

      _other ->
        sse_loop(conn, pid)
    end
  end

  defp format_sse_message(message) do
    "id: #{UUID.uuid1()}\nevent: message\ndata: #{message}\n\n"
  end

  defp get_token(conn) do
    case List.first(get_req_header(conn, "cookie")) do
      nil ->
        nil

      cookie_string ->
        case Regex.run(~r/token=([^;]+)/, cookie_string) do
          [_, token] -> token
          _ -> nil
        end
    end
  end
end
