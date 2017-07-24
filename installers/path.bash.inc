script_link="$( readlink "$BASH_SOURCE" )" || script_link="$BASH_SOURCE"
apparent_minikube_dir="${script_link%/*}"
if [ "$apparent_minikube_dir" == "$script_link" ]; then
  apparent_minikube_dir=.
fi
minikube_dir="$( command cd -P "$apparent_minikube_dir" > /dev/null && pwd -P )"
bin_path="$minikube_dir/bin"
export PATH=$bin_path:$PATH
