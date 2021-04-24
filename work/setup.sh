sudo apt-get update -y && sudo apt-get install -y vim ca-certificates git
sudo apt-get install -y apt-transport-https software-properties-common
sudo sed -i '/AllowAgentForwarding/s/^#//' /etc/ssh/sshd_config

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable"
sudo apt-get update -y && sudo apt install -y docker-ce
sudo usermod -aG docker ${USER}

sudo curl -L https://github.com/docker/compose/releases/download/1.25.5/docker-compose-`uname -s`-`uname -m` -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

wget https://dl.google.com/go/go1.14.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.14.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> .bashrc
go env -w GOPRIVATE="github.com/oligoden"
git config --global url."git@github.com:".insteadOf "https://github.com/"