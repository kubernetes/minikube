build {
  sources = [
    "source.azure-arm.minikube-ci-windows-11"
  ]

  # Deploy SSH public key for local administrator.
  # Required by install-openssh.ps1 to generate C:\ProgramData\ssh\administrators_authorized_keys
  provisioner "powershell" {
    inline = [
      "Write-Host '>>> Copying SSH public key to be deployed for local administrator'",
      "New-Item -Path C:/ProgramData/ssh -ItemType Directory -Force | Out-Null",
      "Add-Content -Path C:/ProgramData/ssh/admin.pub.txt -Value '${var.vm_admin_ssh_public_key}' -Force"
    ]
  }

  ###### <Provisioners with elevated privileges>
  # These provisioners are required for DISM (e.g. install OpenSSH) and other actions,
  # and are implemented in terms scheduled tasks which require AutoLogon to complete.
  provisioner "powershell" {
    script = "provisioners/enable-autologon.ps1"
    environment_vars = [
      "AUTOLOGON_USER_PASSWORD=${var.vm_admin_password}"
    ]
  }

  provisioner "windows-restart" {}

  provisioner "powershell" {
    elevated_user     = var.vm_admin_username
    elevated_password = var.vm_admin_password
    scripts = [
      "provisioners/set-windows.ps1",
      "provisioners/set-privacy.ps1",
      "provisioners/remove-bloatware.ps1",
      "provisioners/remove-services.ps1",
      "provisioners/add-choco.ps1",
      "provisioners/add-openssh.ps1",
      "provisioners/add-hyperv.ps1"
    ]
  }

  provisioner "windows-restart" {}

  provisioner "powershell" {
    elevated_user     = var.vm_admin_username
    elevated_password = var.vm_admin_password
    scripts = [
      "provisioners/add-docker.ps1",
      "provisioners/add-tools.ps1"
    ]
  }

  provisioner "powershell" {
    script = "provisioners/disable-autologon.ps1"
  }
  ###### </Provisioners with elevated privileges>
}
