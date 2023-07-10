defmodule Notification.Nats do
  @moduledoc """
  Nats init connection between process and NATS
  """
  @gnat_process_name :gnat

  require Logger

  def start_subscription(subject) do
    gnat =
      case Process.whereis(@gnat_process_name) do
        nil ->
          Logger.info("Starting NATS...")
          {:ok, gnat} = Gnat.start_link(Application.fetch_env!(:notification, :nats))
          Process.register(gnat, @gnat_process_name)
          gnat

        gnat ->
          gnat
      end

    {:ok, _subscription} = Gnat.sub(gnat, self(), subject)
    receive_messages(subject)
  end

  defp receive_messages(subject) do
    receive do
      {:msg, %{body: body, topic: subject, reply_to: nil}} ->
        PubSub.publish(subject, {body})
        receive_messages(subject)

      _ ->
        receive_messages(subject)
    end
  end
end
