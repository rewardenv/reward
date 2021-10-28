### Reward prerequisites and installation on Ubuntu

#### Uninstall old versions of Docker

Older versions of Docker were called docker, docker.io, or docker-engine. If these are installed, uninstall them:

```
sudo apt-get remove docker docker-engine docker.io containerd runc
```

#### Install using the repository

##### Set up the repository

* Update the apt package index and install packages to allow apt to use a repository over HTTPS:

```
sudo apt-get update
```

```
sudo apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release
```

* Add Docker’s official GPG key:

```
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
```

* Use the following command to set up the stable repository.

```
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

#### Install Docker Engine

* Update the apt package index, and install the latest version of Docker Engine and containerd

```
 sudo apt-get update
 sudo apt-get install docker-ce docker-ce-cli containerd.io
```

* Verify that Docker Engine is installed correctly by running the hello-world image.

```
sudo docker run hello-world
```

* Add your user to docker group to be able to use docker without sudo

```
sudo usermod -aG docker $USER
```

Now you have to **log out and log back** in to apply the user group change.

#### Install Docker Compose

##### Prerequisites

Docker Compose relies on Docker Engine for any meaningful work, so make sure you have Docker Engine installed.
(See the previous step.)

* There are multiple ways to install `docker-compose`, we are going to use python3-pip, so make sure it's installed on
  your system.

```
sudo apt-get install -y python3-pip
```

* Run this command to install the current stable release of Docker Compose using pip:

```
sudo  pip install docker-compose
```

* Test the installation. It should be version >= `1.26`.

```
docker-compose --version
```

#### Installing Reward

* Download the latest package and install it with dpkg.

```
curl -fsSLO "https://github.com/rewardenv/reward/releases/latest/download/reward_`uname -s`_`uname -m`.deb"
sudo dpkg -i "reward_`uname -s`_`uname -m`.deb"
```

```
reward install
```

* Verify the installation

```
reward --version
```
