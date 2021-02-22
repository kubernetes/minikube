Name: minikube
Version: --VERSION--
Release: --REVISION--
Summary: Run Kubernetes locally
License: ASL 2.0
Group: Development/Tools
URL: https://github.com/kubernetes/minikube

# Needed for older versions of RPM
BuildRoot: %{_tmppath}%{name}-buildroot

%description
Minikube is a tool that makes it easy to run Kubernetes locally.
Minikube runs a single-node Kubernetes cluster inside a VM on your
laptop for users looking to try out Kubernetes or develop with it 
day-to-day.

%prep
mkdir -p %{name}-%{version}
cd %{name}-%{version}
cp --OUT--/minikube-linux-%{_arch} minikube

%install
cd %{name}-%{version}
mkdir -p %{buildroot}%{_bindir}
install -m 755 minikube %{buildroot}%{_bindir}/%{name}

%files
%{_bindir}/%{name}
