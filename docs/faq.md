## FAQ

* Can I use on Windows?

  * Yes, Reward is cross-platform. It was written in Golang with intention to support Windows users as well.

* Should I run reward as **root** user or with sudo?

  * Nope, you should almost never use reward as root user or with sudo. The only exception is running
    the `reward self-update` command.

* Is Reward free?

  * Yes, and it's open source as well.

### Frequent errors

* `docker is not running or docker version is too old`

    If you are sure Docker is running on your system, and you keep getting this error, you should check the following:

    * Docker version meets the system requirements mentioned in the
     [Common Requirements](installation.html#common-requirements) section.

    * Your user is not in the `docker` group, and it cannot reach Docker's socket.
        * You can check your users groups with `id` command
        * Also, if you just installed docker, and you just added your user to the docker group you will have to log out
          and log in. For more info go to the following link:
          [Install Docker Engine in Ubuntu](https://docs.docker.com/engine/install/ubuntu/#install-using-the-convenience-script)
          See the `If you would like to use Docker as a non-root user` section.

* `Error: exit status x`

    Most of the cases these errors are coming from the container or docker itself. Reward tries to run a command inside
    the container, and the exit code of the command is not 0, or the container exited during the execution.
    In most cases this error code will be a part of a longer error message which will describe the problem.


* `Error: exit status 137`

    During Magento 2 installation (`reward bootstrap`) you will get this error code most likely during
    the `composer install` command. Composer installation needs a huge amount of memory and this error code
    represents a docker `out of memory` error code.

    To solve this problem, you will have to increase the memory limit of Docker Desktop. For more info see:
    [Additional requirements (macOS only)](installation.html#additional-requirements-macos-only)
