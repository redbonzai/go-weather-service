
* Run server locally with envs:

  ```bash
  make run NWS_USER_AGENT="go-weather-service/1.0 (you@company.com)"
  ```

* Build binary:

  ```bash
  make build
  ```

* Build & run Docker:

  ```bash
  make docker-build
  make docker-run NWS_USER_AGENT="go-weather-service/1.0 (you@company.com)"
  ```

* Compose:

  ```bash
  make compose-up
  # ...
  make compose-down
  ```

