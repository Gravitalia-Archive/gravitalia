FROM elixir:1.15-alpine AS build

WORKDIR /app

RUN mix local.hex --force && \
    mix local.rebar --force

COPY mix.exs ./

COPY lib lib/

RUN mix deps.get

ENV MIX_ENV=prod
RUN mix release

EXPOSE 8891
CMD ["./_build/prod/rel/notification/bin/notification", "start"]