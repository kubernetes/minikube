Vagrant.configure('2') do |config|
    # grab Ubuntu 15.04 official image
    config.vm.box = "ubuntu/vivid64" # Ubuntu 15.04

    # fix issues with slow dns http://serverfault.com/a/595010
    config.vm.provider :virtualbox do |vb, override|
        vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
        # add more ram, the default isn't enough for the build
        vb.customize ["modifyvm", :id, "--memory", "1024"]
    end

    config.vm.synced_folder ".", "/vagrant", type: "rsync"
    config.vm.provision :shell, :privileged => true, :path => "scripts/install-vagrant.sh"
end
