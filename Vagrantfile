Vagrant.configure("2") do |config|
  config.vm.box = "bento/ubuntu-24.04"

  # Read network interface from command-line argument or default to "eth0"
  bridge_adapter = ENV['NETWORK_CARD'] || "eth0"

  machines = [
    "lb-0", "lb-1", "lb-2",
    "node-0", "node-1", "node-2"
  ]

  # Read SSH key name from environment or default to "~/.ssh/vagrant.pub"
  ssh_key_name = ENV['SSH_KEY_PATH'] || "~/.ssh/vagrant.pub"

  # Read the SSH key
  public_key = File.read(File.expand_path(ssh_key_name)).strip

  machines.each do |hostname|
    config.vm.define hostname do |vm|
      vm.vm.hostname = hostname  # Set hostname
      vm.vm.network "public_network", bridge: bridge_adapter  # Use user-specified network card

      vm.vm.provider "virtualbox" do |vb|
        vb.memory = "2048"
        vb.cpus = 2
      end

      # Provisioning: Install Avahi, netcat, and run the small_server.sh script
      vm.vm.provision "shell", inline: <<-SHELL
        sudo apt update
        sudo apt install -y avahi-daemon netcat-traditional

        # Enable and start avahi-daemon
        sudo systemctl enable --now avahi-daemon

        # Add SSH public key login for vagrant user
        sudo mkdir -p /home/vagrant/.ssh
        echo "#{public_key}" | sudo tee -a /home/vagrant/.ssh/authorized_keys > /dev/null
      SHELL

      # Use Vagrant's file provisioner to copy the small_server.sh file to the VM
      vm.vm.provision "file", source: "e2e/small_server.py", destination: "/home/vagrant/small_server.py"
      vm.vm.provision "file", source: "e2e/small_server.service", destination: "/home/vagrant/small_server.service"

      # Run the small_server.sh script
      vm.vm.provision "shell", inline: <<-SHELL
        sudo chmod +x /home/vagrant/small_server.py
        sudo mkdir -p /var/log/small_server

        sudo mv /home/vagrant/small_server.service /etc/systemd/system/small_server.service

        sudo systemctl daemon-reload
        sudo systemctl enable small_server
        sudo systemctl restart small_server
      SHELL
    end
  end
end