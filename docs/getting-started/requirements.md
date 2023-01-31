## Requirements

### Installation Summary

Installing Reward is relatively easy. You can just go to the Reward downloads page and get the binary for your platform.
You should then extract it to any directory and add that directory to your system's PATH.

Buf if you prefer, you can use package managers as well. See in the [Downloading Reward](download.md) section of this
page. But there are some requirements that you need to meet before you can install Reward.

### Common requirements

* **Docker CE 20.10.2** or later
    * [Docker for Linux](https://docs.docker.com/engine/install/#server) (Reward has been tested on Fedora 31 and Ubuntu
      18.04, 20.04).
    * [Docker Desktop for Mac](https://hub.docker.com/editions/community/docker-ce-desktop-mac) 20.10.2 or later.
    * [Docker Desktop for Windows](https://hub.docker.com/editions/community/docker-ce-desktop-windows/) 20.10.2 or
      later.
* **docker-compose version 1.27.4** or later is required (this can be installed via `brew`, `dnf`, or `pip3` as needed).
  It is currently a part of the Docker Desktop package.

    ``` warning::
        **Beware of system package managers!** Some operating system distributions include Docker and docker-compose package in their upstream package repos. Please do not install Docker and docker-compose in this manner. Typically these packages are very outdated versions. If you install via your system's package manager, it is very likely that you will experience issues. Please use the official installers on the Docker Install page.
    ```

    ``` warning::
        If you keep getting the **docker api is unreachable** error message, check the FAQ.
    ```

---

### Additional requirements (macOS only)

* [Mutagen](https://github.com/mutagen-io/mutagen/releases/latest) 0.13.1 or later is required for environments
  leveraging sync sessions on macOS. Reward will attempt to install mutagen via `brew` if not present on macOS.

    ``` warning::
        **By default Docker Desktop allocates 2GB memory.** This leads to extensive swapping, killed processes and extremely high CPU usage during some Magento actions, like for example running sampledata:deploy and/or installing the application. It is recommended to assign at least 6GB RAM to Docker Desktop prior to deploying any Magento environments on Docker Desktop. This can be corrected via Preferences -> Resources -> Advanced -> Memory. While you are there, it wouldn't hurt to let Docker have the use of a few more vCPUs (keep it at least 4 less than the maximum CPU allocation however to avoid having macOS contend with Docker for use of cores)
    ```

---

### Additional requirements (Windows only)

* [Mutagen](https://github.com/mutagen-io/mutagen/releases/latest) 0.13.1 or later is required for environments
  leveraging sync sessions on Windows. Reward will attempt to install mutagen to the same path it is installed.
* [YogaDNS](https://www.yogadns.com/download/) 1.16 Beta or later is required for using dnsmasq as a local DNS resolver
  on Windows.

    ``` warning::
        **On Windows with WSL2 docker can use unlimited memory and CPU.** It is possible and suggested to configure limitations to WSL.
        You can create a .wslconfig file to your user's home directory with the following content.
    
        `C:\\Users\\<yourUserName>\\.wslconfig`
    
        .. code:: ini
    
            [wsl2]
            memory=8GB
            processors=4
    
        If you configured this you will have to restart WSL with the following PowerShell command:
    
        .. code::
    
            Restart-Service LxssManager
    
        See further instructions here: https://docs.microsoft.com/en-us/windows/wsl/wsl-config#wsl-2-settings
    ```
