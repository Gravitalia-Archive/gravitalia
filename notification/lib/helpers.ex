defmodule Notification.Helpers do
  @moduledoc """
  Helpers are a set of functions designed for use in the program
  """

  def check_token(token) do
    IO.puts(Application.get_env(:notification, :jwt))
    {ok, claims} =
      Joken.Signer.verify(
        token,
        Joken.Signer.create("RS256", %{"pem" => Application.get_env(:notification, :jwt)})
      )

    cond do
      ok != :ok ->
        nil

      :os.system_time(:millisecond) / 1000 >= Map.get(claims, "exp") ->
        nil

      true ->
        Map.get(claims, "sub")
    end
  end
end
