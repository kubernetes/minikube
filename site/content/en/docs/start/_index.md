---
Titre: "Commencer Minikube"
Titre lié: "Commencer"
Poids: 1
alias:
  - /docs/start
---
Minikube est un Kubernetes local, qui vise à faciliter l'apprentissage et le développement pour Kubernetes.

Tout ce dont vous avez besoin est un conteneur Docker (ou similaire) ou un environnement de machine virtuelle, et Kubernetes est accessible via  une seule commande: `minikube start`

## Ce dont vous aurez besoin

* 2 processeurs ou plus
* 2 Go de mémoire libre
* 20 Go d'espace de disque libre
* Une connection internet
* Un Gestionnaire de conteneur ou de machine virtuelle, tel que: [Docker]({{<ref "/docs/drivers/docker">}}), [Hyperkit]({{<ref "/docs/drivers/hyperkit">}}), [Hyper-V]({{<ref "/docs/drivers/hyperv">}}), [KVM]({{<ref "/docs/drivers/kvm2">}}), [Parallels]({{<ref "/docs/drivers/parallels">}}), [Podman]({{<ref "/docs/drivers/podman">}}), [VirtualBox]({{<ref "/docs/drivers/virtualbox">}}), or [VMWare]({{<ref "/docs/drivers/vmware">}})

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">1</strong></span>Installation</h2>

{{% card %}}

Cliquez sur les boutons qui correspondent à votre plateforme cible. Pour d'autres architectures, voir [la page de publication] (https://github.com/kubernetes/minikube/releases/latest) pour une liste complète des binaires minikube.

{{% quiz_row base="" name="Système opérateur" %}}
{{% quiz_button option="Linux" %}} {{% quiz_button option="macOS" %}} {{% quiz_button option="Windows" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux" name="Architecture" %}}
{{% quiz_button option="x86-64" %}} {{% quiz_button option="ARM64" %}} {{% quiz_button option="ARMv7" %}} {{% quiz_button option="ppc64" %}} {{% quiz_button option="S390x" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/x86-64" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}} {{% quiz_button option="Debian package" %}} {{% quiz_button option="RPM package" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/ARM64" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}} {{% quiz_button option="Debian package" %}} {{% quiz_button option="RPM package" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/ppc64" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/S390x" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/ARMv7" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/macOS" name="Architecture" %}}
{{% quiz_button option="x86-64" %}} {{% quiz_button option="ARM64" %}}
{{% /quiz_row %}}

{{% quiz_row base="/macOS/x86-64" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}} {{% quiz_button option="Homebrew" %}}
{{% /quiz_row %}}

{{% quiz_row base="/macOS/ARM64" name="Type d'installation" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Windows" name="Architecture" %}}
{{% quiz_button option="x86-64" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Windows/x86-64" name="Type d'installation" %}}
{{% quiz_button option=".exe download" %}} {{% quiz_button option="Windows Package Manager" %}} {{% quiz_button option="Chocolatey" %}}
{{% /quiz_row %}}

{{% quiz_instruction id="/Linux/x86-64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/x86-64/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_amd64.deb
sudo dpkg -i minikube_latest_amd64.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/x86-64/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.x86_64.rpm
sudo rpm -Uvh minikube-latest.x86_64.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARM64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-arm64
sudo install minikube-linux-arm64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARM64/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_arm64.deb
sudo dpkg -i minikube_latest_arm64.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARM64/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.aarch64.rpm
sudo rpm -Uvh minikube-latest.aarch64.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ppc64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-ppc64le
sudo install minikube-linux-ppc64le /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ppc64/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_ppc64le.deb
sudo dpkg -i minikube_latest_ppc64le.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ppc64/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.ppc64el.rpm
sudo rpm -Uvh minikube-latest.ppc64el.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/S390x/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-s390x
sudo install minikube-linux-s390x /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/S390x/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_s390x.deb
sudo dpkg -i minikube_latest_s390x.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/S390x/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.s390x.rpm
sudo rpm -Uvh minikube-latest.s390x.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARMv7/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-arm
sudo install minikube-linux-arm /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/macOS/x86-64/Homebrew" %}}
Si le [Brew Package Manager] (https://brew.sh/) est installé:

```shell
brew install minikube
```


Si `which minikube` échoue après l'installation via brew, vous devrez peut-être supprimer les anciens liens minikube et lier le binaire nouvellement installé:

```shell
brew unlink minikube
brew link minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/macOS/x86-64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64
sudo install minikube-darwin-amd64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/macOS/ARM64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-arm64
sudo install minikube-darwin-arm64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Windows/x86-64/Windows Package Manager" %}}
Si le [Gestionnaire de packages Windows] (https://docs.microsoft.com/en-us/windows/package-manager/) est installé, utilisez la commande suivante pour installer minikube:

```shell
winget install minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Windows/x86-64/Chocolatey" %}}
Si le [Chocolatey Package Manager] (https://chocolatey.org/) est installé, utilisez la commande suivante:

```shell
choco install minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Windows/x86-64/.exe download" %}}
Téléchargez et exécutez le [programme d'installation Windows minikube] (https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe).

_Si vous avez utilisé une CLI pour effectuer l'installation, vous devrez fermer cette CLI et en ouvrir une nouvelle avant de continuer ._
{{% /quiz_instruction %}}

{{% /card %}}


<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">2</strong></span>Start your cluster</h2>

Depuis un terminal avec un accès administrateur (mais non connecté en tant que root), exécutez:

```shell
minikube start
```

Si minikube ne démarre pas, consultez la [page des pilotes] ({{<ref "/ docs / drivers">}}) pour obtenir de l'aide sur la configuration d'un conteneur compatible ou d'un gestionnaire de machine virtuelle.

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">3</strong></span>Interact with your cluster</h2>

Si vous avez déjà installé kubectl, vous pouvez maintenant l'utiliser pour accéder à votre nouveau cluster :

```shell
kubectl get po -A
```

Sinon, minikube peut télécharger la version appropriée de kubectl, si cela ne vous dérange pas d'utiliser des doubles tirets dans la ligne de commande:

```shell
minikube kubectl -- get po -A
```

Au départ, certains services tels que le fournisseur de stockage ne sont peut-être pas encore en cours d'exécution. Il s'agit d'une condition normale lors de la mise en place du cluster et se résoudra d'elle-même momentanément. Pour plus d'informations sur l'état de votre cluster, minikube possède le tableau de bord Kubernetes, vous permettant de vous acclimater facilement à votre nouvel environnement:

```shell
minikube dashboard
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">4</strong></span>Déployer une application</h2>

Créez un exemple de déploiement et exposez-le sur le port 8080:

```shell
kubectl create deployment hello-minikube --image=k8s.gcr.io/echoserver:1.4
kubectl expose deployment hello-minikube --type=NodePort --port=8080
```

Cela peut prendre un moment, mais votre déploiement apparaîtra bientôt lorsque vous exécuterez:

```shell
kubectl get services hello-minikube
```

Le moyen le plus simple d'accéder à ce service est de laisser minikube lancer un navigateur Web pour vous:

```shell
minikube service hello-minikube
```


Vous pouvez également utiliser kubectl pour transférer le port:

```shell
kubectl port-forward service/hello-minikube 7080:8080
```

Tada! Votre application est désormais disponible à l'adresse [http://localhost:7080/] (http://localhost:7080/)

### Déploiements de charge équilibrée

Pour accéder à un déploiement de charge équilibrée, utilisez la commande "minikube tunnel". Voici un exemple de déploiement:

```shell
kubectl create deployment balanced --image=k8s.gcr.io/echoserver:1.4  
kubectl expose deployment balanced --type=LoadBalancer --port=8080
```

Dans une autre fenêtre, démarrez le tunnel pour créer une IP routable pour le déploiement ``équilibré'':

```shell
minikube tunnel
```

Pour trouver l'adresse IP routable, exécutez cette commande et examinez la colonne `EXTERNAL-IP`:

```shell
kubectl get services balanced
```

Votre déploiement est désormais disponible à l'adresse  &lt;EXTERNAL-IP&gt;:8080

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">5</strong></span>Gérer votre cluster</h2>

Suspendez Kubernetes sans affecter les applications déployées:

```shell
minikube pause
```

Arrêtez le cluster:

```shell
minikube stop
```

Augmentez la limite de mémoire par défaut (nécessite un redémarrage):

```shell
minikube config set memory 16384
```

Parcourez le catalogue de services Kubernetes faciles à installer:

```shell
minikube addons list
```

Créez un deuxième cluster exécutant une ancienne version de Kubernetes:

```shell
minikube start -p aged --kubernetes-version=v1.16.1
```

Supprimez tous les clusters de minikube:

```shell
minikube delete --all
```

## Passez à l'étape suivante

* [Le manuel du minikube]({{<ref "/docs/handbook">}})
* [Tutoriels fournis par la communauté]({{<ref "/docs/tutorials">}})
* [Référence pour la commande minikube]({{<ref "/docs/commands">}})
* [Guide des contributeurs]({{<ref "/docs/contrib">}})
* Remplissez notre [enquête rapide en 5 questions](https://forms.gle/Gg3hG5ZySw8c1C24A) pour partager vos ressentis 🙏
