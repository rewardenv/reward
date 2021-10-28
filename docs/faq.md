## FAQ

* Can I use on Windows?

    * Yes, Reward is cross-platform. It was written in Golang with intention to support Windows users as well.

* Should I run reward as **root** user or with sudo?

    * Nope, you should almost never use reward as root user or with sudo. The only exception is running
      the `reward self-update` command.

* Is Reward free?

    * Yes, and it's open source as well.

* Can I connect to the database using root user?

    * Yes, run `reward db connect --root`.

### Frequent errors

* `docker api is unreachable`

  If you are sure Docker is running on your system, and you keep getting this error, you should check the following:

    * **Make sure your Docker version is up to date** and meets the system requirements mentioned in the
      [Common Requirements](installation.html#common-requirements) section.
    ``` note::
           Package managers provide outdated Docker versions.
    ```

    * **Your user is not in the `docker` group**, or it cannot reach the docker socket.
        * After you add your user to the docker group ***you will have to reboot*** (or log out and log back in). For
          more info go to the following link:
          [Install Docker Engine in Ubuntu](https://docs.docker.com/engine/install/ubuntu/#install-using-the-convenience-script)
          See the `If you would like to use Docker as a non-root user` section.

    ``` note::
           You can check if your user is in the docker group with **id** command.
  
           You can make sure your user is able to reach the docker API running **docker ps** (without sudo).
    ```

---

* `Error: exit status x`

  Most of the cases these errors are coming from the container or docker itself. Reward tries to run a command inside
  the container, and the exit code of the command is not 0, or the container exited during the execution. In most cases
  this error code will be a part of a longer error message which will describe the problem.

---

* `Error: exit status 137`

  During Magento 2 installation (`reward bootstrap`) you will get this error code most likely during
  the `composer install` command. Composer installation needs a huge amount of memory and this error code represents a
  docker `out of memory` error code.

  To solve this problem, you will have to increase the memory limit of Docker Desktop. For more info see:
  [Additional requirements (macOS only)](installation.html#additional-requirements-macos-only)

---

* ```Error: unable to connect to beta: unable to connect to endpoint: unable to dial agent endpoint: unable to install agent: unable to get agent for platform: unable to locate agent bundle```

  This error message most probably appears if Mutagen is installed on your system but the Mutagen Agents Bundle are
  not. **Check your PATH (like `c:\bin`) and make sure `mutagen.exe` and `mutagen-agents.tar.gz` are there**. If not,
  download the latest release of Mutagen and extract the archive. You will find both files in it.

---

* ```Error: create failed: unable to connect to beta: unable to connect to endpoint: unable to dial agent endpoint: unable to create agent command: unable to probe container: container probing failed under POSIX hypothesis (signal: killed) and Windows hypothesis (signal: killed)```

  If this error occurs, you can try to install mutagen-beta.
    ```
    mutagen daemon stop
    brew uninstall mutagen
    brew install mutagen-io/mutagen/mutagen-beta
    ```

---

* ```Error: unable to connect to daemon: client/daemon version mismatch (daemon restart recommended)```

  There's a possibility mutagen was updated on your system. Homebrew update doesn't restart mutagen daemon, try to run
  it manually:
    ```
    mutagen daemon stop
    mutagen daemon start
    ```

---

* `Package hirak/prestissimo has a PHP requirement incompatible with your PHP version, PHP extensions and Composer version`

  If you see this error message during the Magento 2 installation, you will have to downgrade your Composer version.

  To do so, add the following line to the `.env`:
    ```
    COMPOSER_VERSION=1
    ```

  For more information, see the [Composer configuration](configuration/composer.md).

---

* `reward shell` stuck or frozen on Windows

  If any of Reward commands seems to be frozen or stuck, and you are using Git Bash on Windows, most probably you faced
  a known issue with Docker for Windows and Git Bash. To solve it, you have to use `winpty` (Windows Pseudo Terminal)
  using the following command:
    ```
    winpty reward shell
    ```
  If you are interested in what's happening in the background, you can find more information here:
    * https://github.com/docker/for-win/issues/1588
    * https://willi.am/blog/2016/08/08/docker-for-windows-interactive-sessions-in-mintty-git-bash/