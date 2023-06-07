## Setting up Reward

### Install Reward

Before you first run Reward, you will have to install Reward configurations.

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

---

### Provision Reward's Global Services

After you installed Reward's basic settings you'll be able to provision the global services such as Traefik, Portainer
and so on.

To do that, run the following command:

``` shell
$ reward svc up
```

For more information check the [Global Services](../services.md) guide.

---

### Automatic DNS Resolution

#### Linux

On Linux environments, Reward tries to configure your NetworkManager and systemd-resolved to use the local resolver. If
it's not working, you will have to configure your DNS to resolve `*.test` to `127.0.0.1` or use `/etc/hosts` entries.

After Reward is installed, probably you will have to restart your NetworkManager (or reboot your system).

For further information, see the [Automatic DNS Resolution](../configuration/automatic-dns-resolution.html#linux) guide.

#### macOS

This configuration is automatic via the BSD per-TLD resolver configuration found at `/etc/resolver/test`.

#### Windows

We suggest to use [YogaDNS](https://www.yogadns.com/download/) which allows you to create per domain rules for DNS
resolution. With YogaDNS you can configure your OS to ask dnsmasq for all `*.test` domain and use your default Name
Server for the rest.

For more information see the configuration page
for [Automatic DNS Resolution](../configuration/automatic-dns-resolution.html#windows)

---

### Trusted CA Root Certificate

In order to sign SSL certificates that may be trusted by a developer workstation, Reward uses a CA root certificate with
CN equal to `Reward Proxy Local CA (<hostname>)` where `<hostname>` is the hostname of the machine the certificate was
generated on at the time Reward was first installed. The CA root can be found
at `~/.reward/ssl/rootca/certs/ca.cert.pem`.

#### Linux

On Ubuntu/Debian this CA root is copied into `/usr/local/share/ca-certificates` and on Fedora/CentOS (Enterprise Linux)
it is copied into `/etc/pki/ca-trust/source/anchors` and then the trust bundle is updated appropriately. For new
systems, this typically is all that is needed for the CA root to be trusted on the default Firefox browser, but it may
not be trusted by Chrome or Firefox automatically should the browsers have already been launched prior to the
installation of Reward (browsers on Linux may and do cache CA bundles)

#### macOS

On macOS this root CA certificate is automatically added to a users trust settings as can be seen by searching for '
Reward Proxy Local CA' in the Keychain application. This should result in the certificates signed by Reward being
trusted by Safari and Chrome automatically. If you use Firefox, you will need to add this CA root to trust settings
specific to the Firefox browser per the below.

#### Windows

On Windows this root CA certificate is automatically added to the users trust settings as can be seen by searching for '
Reward Proxy Local CA' in the Management Console. This should result in the certificates signed by Reward being trusted
by Edge, Chrome and Firefox automatically.

``` note::
    If you are using **Firefox** and it warns you the SSL certificate is invalid/untrusted, go to Preferences -> Privacy & Security -> View Certificates (bottom of page) -> Authorities -> Import and select ``~/.reward/ssl/rootca/certs/ca.cert.pem`` for import and make sure you select **'Trust this CA to identify websites'**. Then reload the page.

    If you are using **Chrome** on **Linux** and it warns you the SSL certificate is invalid/untrusted, go to Chrome Settings -> Privacy And Security -> Manage Certificates (see more) -> Authorities -> Import and select ``~/.reward/ssl/rootca/certs/ca.cert.pem`` for import and make sure you select **'Trust this certificate for identifying websites'**. Then reload the page.
```

