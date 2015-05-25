# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.network "forwarded_port", guest: 8080, host: 8080

$script = <<SCRIPT
    sudo apt-get update
    sudo apt-get upgrade
    sudo apt-get install -y git 

    # install  golang
    wget --quiet https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
    ln -s /usr/local/go/bin/* /usr/bin/

    # install  mongodb
    sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
    echo "deb http://repo.mongodb.org/apt/ubuntu trusty/mongodb-org/3.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.0.list
    sudo apt-get update
    sudo apt-get install -y mongodb-org

   # configure go
   mkdir ~/go
   cd /vagrant
   GOPATH=~/go go get -t
   GOPATH=~/go go build -o lsmsd
SCRIPT

  config.vm.provision "shell", inline: $script

end
