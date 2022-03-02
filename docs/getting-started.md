## Getting Started

### Configuring Reward

When you first run Reward, you will have to install Reward configurations.

``` shell
$ reward install
```

This is going to do the following:

* Create a Self Signed Root CA Certificate and install it to you operating system's Root CA Trust. To do so Reward will
  ask for your sudo / administrator permission.
* Configure your Operating System's DNS resolver to use Reward's dnsmasq service to resolve *.test domains (macOS and
  Linux only).
* Create an SSH Tunnel Key and configure SSH (/etc/ssh/ssh_config) to use this key if you want to utilize Reward's
  tunnel (macOS and Linux only).

### Provision Reward's Global Services

After you installed Reward's basic settings you'll be able to provision the global services such as Traefik, Portainer
and so on.

To do that, run the following command:

``` shell
$ reward svc up
```
