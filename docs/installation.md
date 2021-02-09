## Installation

### Summary

Installing Reward is relatively easy. You can just go to the Reward downloads page and get the binary for your platform.
You should then extract it to any directory and add that directory to your system's PATH.

Buf if you prefer, you can use package managers as well. See in the [Installing Reward](installation.html#installing-reward) section of this page.

### Prerequisites

* Docker
* docker-compose
* mutagen for syncing (Windows and macOS only)
* YogaDNS for local DNS resolution (Windows only)

#### Details
##### Common requirements

* Docker CE 20.10.2 or later
  * [Docker for Linux](https://docs.docker.com/install/) (Reward has been tested on Fedora 31 and Ubuntu 18.04, 20.04).
  * [Docker Desktop for Mac](https://hub.docker.com/editions/community/docker-ce-desktop-mac) 20.10.2 or later.
  * [Docker Desktop for Windows](https://hub.docker.com/editions/community/docker-ce-desktop-windows/) 20.10.2 or later.
* docker-compose version 1.27.4 or later is required (this can be installed via `brew`, `apt`, `dnf`, or `pip3` as needed). It is currently a part of the Docker Desktop package.

##### Additional requirements (macOS only)

* [Mutagen](https://github.com/mutagen-io/mutagen/releases/) 0.11.8 or later is required for environments leveraging sync sessions on macOS. Reward will attempt to install mutagen via `brew` if not present on macOS.

``` warning::
    **By default Docker Desktop allocates 2GB memory.** This leads to extensive swapping, killed processes and extremely high CPU usage during some Magento actions, like for example running sampledata:deploy and/or installing the application. It is recommended to assign at least 6GB RAM to Docker Desktop prior to deploying any Magento environments on Docker Desktop. This can be corrected via Preferences -> Resources -> Advanced -> Memory. While you are there, it wouldn't hurt to let Docker have the use of a few more vCPUs (keep it at least 4 less than the maximum CPU allocation however to avoid having macOS contend with Docker for use of cores)
```

##### Additional requirements (Windows only)

* [Mutagen](https://github.com/mutagen-io/mutagen/releases/) 0.11.8 or later is required for environments leveraging sync sessions on Windows. Reward will attempt to install mutagen to the same path it is installed.
* [YogaDNS](https://www.yogadns.com/download/) 1.16 Beta or later is required for using dnsmasq as a local DNS resolver on Windows.

``` warning::
    **On Windows with WSL2 docker can use unlimited memory and CPU.** It is possible and suggested to configure limitations to WSL.
    You can create a .wslconfig file to your user's home directory with the following content.

    `C:\\Users\\<yourUserName>\\.wslconfig`

    .. code:: ini

        [wsl2]
        memory=8GB
        processors=4

    See further instructions here: https://docs.microsoft.com/en-us/windows/wsl/wsl-config#wsl-2-settings
```

### Installing Reward

#### Linux

##### Ubuntu

```
$ curl -sS -O -L "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.deb"
$ sudo dpkg -i "reward_`uname -s`_`uname -m`.deb"
```

##### CentOS and Fedora

```
$ yum install -y "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.rpm"
```

##### Binary Download

```
$ curl -sS -O -L "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.tar.gz"
$ tar -zxvf "reward_`uname -s`_`uname -m`.tar.gz" -C /usr/local/bin/
$ rm -f "reward_`uname -s`_`uname -m`.tar.gz"
$ chmod +x /usr/local/bin/reward
```

#### macOS

You can install reward using Homebrew or by downloading the binary itself and putting it to PATH.

##### Using Homebrew
```
$ brew install rewardenv/tap/reward
```

##### Binary download
```
$ curl -sS -O -L "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.tar.gz"
$ tar -zxvf "reward_`uname -s`_`uname -m`.tar.gz" -C /usr/local/bin/
$ rm -f "reward_`uname -s`_`uname -m`.tar.gz"
$ chmod +x /usr/local/bin/reward
```

#### Windows

[Download Reward from this link](https://github.com/rewardenv/reward/releases/latest/download/reward_windows_x86_64.zip)
and extract to any folder like `C:\bin`. Please make sure that folder is in you PATH environment variable.

You can find a nice guide [here](https://www.architectryan.com/2018/03/17/add-to-the-path-on-windows-10/) about how
to configure PATH in Windows.

### Next Steps

You will have to run `reward install` to initialize Reward. See more in [Getting Started](getting-started.md)

#### Automatic DNS Resolution

##### Linux

On Linux environments, Reward tries to configure your NetworkManager and systemd-resolved to use the local resolver. If it's not working, you will have to configure your DNS to resolve `*.test` to `127.0.0.1` or use `/etc/hosts` entries.

##### macOS

This configuration is automatic via the BSD per-TLD resolver configuration found at `/etc/resolver/test`.

##### Windows

We suggest to use [YogaDNS](https://www.yogadns.com/download/) which allows you to create per domain rules for DNS resolution. With YogaDNS you can configure your OS to ask dnsmasq for all `*.test` domain and use your default Name Server for the rest.

For more information see the configuration page for [Automatic DNS Resolution](configuration/dns-resolver.html#windows)

#### Trusted CA Root Certificate

In order to sign SSL certificates that may be trusted by a developer workstation, Reward uses a CA root certificate with CN equal to `Reward Proxy Local CA (<hostname>)` where `<hostname>` is the hostname of the machine the certificate was generated on at the time Reward was first installed. The CA root can be found at `~/.reward/ssl/rootca/certs/ca.cert.pem`.

##### Linux

On Ubuntu/Debian this CA root is copied into `/usr/local/share/ca-certificates` and on Fedora/CentOS (Enterprise Linux) it is copied into `/etc/pki/ca-trust/source/anchors` and then the trust bundle is updated appropriately. For new systems, this typically is all that is needed for the CA root to be trusted on the default Firefox browser, but it may not be trusted by Chrome or Firefox automatically should the browsers have already been launched prior to the installation of Reward (browsers on Linux may and do cache CA bundles)

##### macOS

On macOS this root CA certificate is automatically added to a users trust settings as can be seen by searching for 'Reward Proxy Local CA' in the Keychain application. This should result in the certificates signed by Reward being trusted by Safari and Chrome automatically. If you use Firefox, you will need to add this CA root to trust settings specific to the Firefox browser per the below.

##### Windows

On Windows this root CA certificate is automatically added to the users trust settings as can be seen by searching for 'Reward Proxy Local CA' in the Management Console. This should result in the certificates signed by Reward being trusted by Edge, Chrome and Firefox automatically.

``` note::
    If you are using **Firefox** and it warns you the SSL certificate is invalid/untrusted, go to Preferences -> Privacy & Security -> View Certificates (bottom of page) -> Authorities -> Import and select ``~/.reward/ssl/rootca/certs/ca.cert.pem`` for import, then reload the page.

    If you are using **Chrome** on **Linux** and it warns you the SSL certificate is invalid/untrusted, go to Chrome Settings -> Privacy And Security -> Manage Certificates (see more) -> Authorities -> Import and select ``~/.reward/ssl/rootca/certs/ca.cert.pem`` for import, then reload the page.
```

### Updating Reward

When Reward is already installed on your system you can do a self-update running `reward self-update` command.

If you installed it using a package manager on linux,
you will have to run it as superuser with `sudo reward self-update`.
