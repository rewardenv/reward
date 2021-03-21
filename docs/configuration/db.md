## Database Connections

### Common Settings

| Name                      | Value/Description                                                     |
| ------------------------- |---------------------------------------------------------------------- |
| MySQL Host                | Name of your Docker Container, can be found with `reward env ps db`   |
| MySQL Port                | `3306`                                                                |
| MySQL User                | `magento`                                                             |
| MySQL Password            | `magento`                                                             |
| MySQL Database            | `magento`                                                             |
| SSH Tunnel Host           | `tunnel.reward.test`                                                  |
| SSH Tunnel Port           | `2222`                                                                |
| SSH Tunnel User           | `user`                                                                |
| SSH Tunnel Key File       | *macOS or Linux* `~/.reward/tunnel/ssh_key`                           |
|                           | *Windows*         `C:\Users\user\.reward\tunnel\ssh_key`              |

``` ...note::
    On Windows use "Key pair Authentication" and select the SSH private key from
    Reward's home directory. See the example above.
```

### TablePlus
![TablePlus Connection Info](screenshots/tableplus-connection.png)

### Sequel Pro / Sequel Ace
![Sequel Pro Connection Info](screenshots/sequel-pro-connection.png)

### PhpStorm
![PHPStorm Connection Config](screenshots/phpstorm-connection-config.png)
![PHPStorm Tunnel Config](screenshots/phpstorm-tunnel-config.png)

### Navicat for MySQL
![Navicat Connection Config](screenshots/navicat-connection-config.png)
![Navicat Tunnel Config](screenshots/navicat-ssh-tunnel-config.png)
