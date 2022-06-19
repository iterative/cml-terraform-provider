#/bin/sh
FILE=/var/log/cml_stack.log
if [ ! -f "$FILE" ]; then
  DEBIAN_FRONTEND=noninteractive
  PS4='tpi:setup.sh: '
  set -x
  echo "APT::Get::Assume-Yes \"true\";" | sudo tee -a /etc/apt/apt.conf.d/90assumeyes

  sudo apt remove unattended-upgrades
  systemctl disable apt-daily-upgrade.service

  sudo add-apt-repository universe -y
  sudo add-apt-repository ppa:git-core/ppa -y
  sudo apt update && sudo apt-get install -y software-properties-common build-essential git

  sudo curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh
  sudo usermod -aG docker ubuntu
  sudo setfacl --modify user:ubuntu:rw /var/run/docker.sock

  get_ecr_helper="curl https://amazon-ecr-credential-helper-releases.s3.us-east-2.amazonaws.com/0.5.0/linux-amd64/docker-credential-ecr-login --output /usr/bin/docker-credential-ecr-login"
  chmod_ecr_help="chmod 755 /usr/bin/docker-credential-ecr-login"
  sudo systemd-run --same-dir --no-block --service-type=exec bash -c "$get_ecr_help && $chmod_ecr_help"

  curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
  sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
  sudo apt update && sudo apt-get install -y terraform

  curl -sL https://deb.nodesource.com/setup_16.x | sudo bash
  sudo apt update && sudo apt-get install -y nodejs

  sudo apt install -y ubuntu-drivers-common
  if ubuntu-drivers devices | grep 'NVIDIA' > /dev/null; then 
    sudo ubuntu-drivers install
    
    curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
    curl -s -L https://nvidia.github.io/nvidia-docker/ubuntu18.04/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list
    sudo apt update && sudo apt install -y nvidia-docker2
    sudo systemctl restart docker
  fi
  
  echo OK | sudo tee "$FILE"
fi
