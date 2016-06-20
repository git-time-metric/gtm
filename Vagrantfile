# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "hashicorp/precise64"

  config.vm.provision "fix-no-tty", type: "shell" do |s|
      s.privileged = false
      s.inline = "sudo sed -i '/tty/!s/mesg n/tty -s \\&\\& mesg n/' /root/.profile"
  end

  config.vm.provision "shell", inline: <<-SHELL
    # install go build environment
    if [[ ! -f "/usr/local/go/bin/go" ]]; then
      sudo apt-get update
      sudo apt-get install -y git curl make cmake pkg-config
      sudo curl -s -O https://storage.googleapis.com/golang/go1.7beta2.linux-amd64.tar.gz
      sudo tar -xvf go1.7beta2.linux-amd64.tar.gz
      sudo mv go /usr/local
      sudo echo "\nPATH=$PATH:/usr/local/go/bin:/home/vagrant/dev/bin\nexport GOPATH=/home/vagrant/dev\n" >> /home/vagrant/.profile

      # setup gtm on GOPATH
      mkdir -p /home/vagrant/dev/src/github.com/git-time-metric
      sudo chown -R vagrant:vagrant /home/vagrant/dev
      sudo ln -s /vagrant /home/vagrant/dev/src/github.com/git-time-metric/gtm
    fi

    echo "Installing git2go and libgit2"

    source /home/vagrant/.profile
    sudo chown -R vagrant:vagrant /home/vagrant/dev
    go get -d github.com/libgit2/git2go
    cd $GOPATH/src/github.com/libgit2/git2go
    git checkout next
    git submodule update --init
    make install

    # test and install gtm
    cd $GOPATH/src/github.com/git-time-metric/gtm
    go get -u github.com/FiloSottile/gvt
    gvt restore
    go test ./...
    go install
    sudo chown -R vagrant:vagrant /home/vagrant/dev
  SHELL

end
