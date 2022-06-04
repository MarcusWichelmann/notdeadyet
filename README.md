# notdeadyet

This is a simple dead man's switch monitoring daemon written in Go. It watches, if an API endpoint is polled at a regular, configurable interval, and triggers an alarm otherwise.

For now, only **Pushover** receivers are supported, but more can be added easily.

## Usage

### Configuration file

Take a look at the following sample to write your own `config.yml`:

```yaml
listen: ':80'  # Optional
apps:
  - name: Test App
    token: djms90x1hqflyggx  # Generate a random token yourself
    timeout: 1h
    repeat_interval: 1h
    notify:
      - Super Administrator
receivers:
  pushover:
    - name: Super Administrator
      user_key: ...
      token: ...
      priority: 0
```

### Run

You can then run the daemon with
```
./notdeadyet --config-file=config.yml
```

or use the provided Docker images and mount your config into `/config/config.yml`.

### Making requests

You should configure your monitored applications to make regular `GET` or `POST` requests against `http://SERVER/im-alive/TOKEN` to tell `notdeadyet` that the app is still alive.

If the monitored app stops making these requests for the duration configured in `timeout`, then the selected receivers will receive a notification. When the app starts making requests again, a notification is sent, that the app is back.

**Have fun!**