## Downloading Reward

### Linux

#### Ubuntu

[Step by Step installation on Ubuntu](install-on-ubuntu.md)

```
$ curl -fsSLO "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.deb"
$ sudo dpkg -i "reward_`uname -s`_`uname -m`.deb"
```

#### CentOS and Fedora

```
$ yum install -y "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.rpm"
```

#### Binary Download

```
$ curl -fsSLO "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.tar.gz"
$ tar -zxvf "reward_`uname -s`_`uname -m`.tar.gz" -C /usr/local/bin/
$ rm -f "reward_`uname -s`_`uname -m`.tar.gz"
$ chmod +x /usr/local/bin/reward
```

---

### macOS

You can install reward using Homebrew or by downloading the binary itself and putting it to PATH.

#### Using Homebrew

```
$ brew install rewardenv/tap/reward
```

#### Binary download

```
$ curl -fsSLO "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.tar.gz"
$ tar -zxvf "reward_`uname -s`_`uname -m`.tar.gz" -C /usr/local/bin/
$ rm -f "reward_`uname -s`_`uname -m`.tar.gz"
$ chmod +x /usr/local/bin/reward
```

---

### Windows

Download Reward [from this link](https://github.com/rewardenv/reward/releases/latest/download/reward_windows_x86_64.zip)
and extract to any folder like `C:\bin`. Please make sure that folder is in you PATH environment variable.

You can find a nice guide [here](https://www.architectryan.com/2018/03/17/add-to-the-path-on-windows-10/) about how to
configure PATH in Windows.

---

### Updating Reward

When Reward is already installed on your system, you can do a self-update running `reward self-update` command.

If you installed it using a package manager on linux, you will have to run it as superuser
with `sudo reward self-update`.

### Next Steps

You will have to run `reward install` to initialize Reward. See more in the [Setting Up](setup.md)
