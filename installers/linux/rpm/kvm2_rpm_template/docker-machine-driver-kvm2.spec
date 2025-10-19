Name: docker-machine-driver-kvm2
Version: --VERSION--
Release: 0
Summary: Machine driver for KVM
License: ASL 2.0
Group: Development/Tools
URL: https://github.com/kubernetes/minikube
#Requires: <determined automatically by rpm>

# Needed for older versions of RPM
BuildRoot: %{_tmppath}%{name}-buildroot

%description
Minikube uses Docker Machine to manage the Kubernetes VM so it benefits
from the driver plugin architecture that Docker Machine uses to provide
a consistent way to manage various VM providers.

%prep
mkdir -p %{name}-%{version}
cd %{name}-%{version}
cp --OUT--/docker-machine-driver-kvm2-%{_arch} docker-machine-driver-kvm2

%install
cd %{name}-%{version}
mkdir -p %{buildroot}%{_bindir}
install -m 755 docker-machine-driver-kvm2 %{buildroot}%{_bindir}/%{name}

%files
%{_bindir}/%{name}
