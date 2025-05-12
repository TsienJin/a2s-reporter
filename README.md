# A2S Prometheus Reporter

[//]: # ([![Docker Pulls]&#40;https://img.shields.io/docker/pulls/your-dockerhub-username/a2s-exporter.svg&#41;]&#40;https://hub.docker.com/r/your-dockerhub-username/a2s-exporter&#41; <!-- Optional: Replace with your Docker Hub link -->)

[//]: # (<!-- Add other badges if you have them &#40;build status, license, etc.&#41; -->)

A simple, lightweight Dockerized Prometheus reporter written in Go. It queries a specified game server using the Valve A2S protocol and exposes key server metrics via an HTTP endpoint (`/metrics`) for Prometheus consumption.

## Features

*   Queries game servers supporting the A2S protocol.
*   Exposes metrics in Prometheus format at the `/metrics` endpoint.
*   Highly configurable via environment variables.
*   Deployable as a Docker container.
*   Minimal resource footprint.

## Configuration

The exporter is configured using environment variables:

| Variable              | Required? | Default                 | Description                                                                                                |
| :-------------------- | :-------- |:------------------------| :--------------------------------------------------------------------------------------------------------- |
| `REPORTER_PORT`       | **Yes**   | `3000`                  | The TCP port the exporter will listen on for Prometheus scrapes (e.g., `/metrics`).                         |
| `GAME_A2S_ADDRESS`    | **Yes**   | *None*                  | The hostname or IP address of the game server to query.                                                    |
| `GAME_A2S_PORT`       | **Yes**   | *None*                  | The UDP port the game server is listening on for A2S queries (often game port or game port + 1).           |
| `DNS_SERVER`          | No        | `host.docker.internal`¹ | (Optional) Specify a custom DNS server IP for the container to use for resolving `GAME_A2S_ADDRESS`.       |
| `QUERY_INTERVAL`      | No        | `10000`                 | Interval between A2S queries in milliseconds (ms).                                                          |
| `QUERY_TIMEOUT`       | No        | `3000`                  | Timeout for waiting for an A2S response in milliseconds (ms).                                               |
| `QUERY_MAX_PACKET_SIZE`| No        | `1400`                  | Some engine does not follow the protocol spec, and may require bigger packet buffer. |

¹ *Note on `DNS_SERVER`*: If specified via the `dns:` directive in `docker-compose.yml`, that takes precedence. `host.docker.internal` is a special Docker DNS name resolving to the host's IP, useful for local DNS resolvers or VPNs like Tailscale.

## Exposed Metrics

The exporter exposes the following metrics at the `/metrics` endpoint:

| Metric Name                           | Type  | Labels                     | Description                                               |
| :------------------------------------ | :---- | :------------------------- | :-------------------------------------------------------- |
| `a2s_server_status`                   | Gauge |                            | 1 if the server responded successfully to the query, 0 if not. |
| `a2s_server_player_count`             | Gauge |                            | Current number of human players on the server.            |
| `a2s_server_player_count_with_server_name` | Gauge | `server`                   | Current player count labelled with the server name.      |
| `a2s_server_max_player_count`         | Gauge |                            | Maximum number of players allowed by the server.          |
| `a2s_server_bots`                     | Gauge |                            | Number of bots (AI players) on the server.                  |
| `a2s_server_map_info`                 | Gauge | `server_name`, `map`       | Value is 1, contains current map and server name as labels. |
| `a2s_server_password_set`             | Gauge |                            | 1 if the server is password protected, 0 if public.       |
| `a2s_server_vac_enabled`              | Gauge |                            | 1 if Valve Anti-Cheat (VAC) is enabled, 0 if not.         |

*Note: Go runtime metrics (`go_*`) and process metrics (`process_*`) are automatically excluded for a cleaner output focused only on A2S data.* are automatically excluded for a cleaner output focused only on A2S data.*