
[Docsy](https://github.com/google/docsy) is a Hugo theme for technical documentation sites, providing easy site navigation, structure, and more. This **Minikube project** uses the Docsy theme for [Minikube Website](https://minikube.sigs.k8s.io/docs/).

You can find detailed theme instructions in the Docsy user guide: https://docsydocs.netlify.com/docs/


## Running the website locally

Clone the minikube project fork with option ```--recurse-submodules --depth 1 ``` to download  and update submodule dependencies.
```bash
git clone --recurse-submodules --depth 1 https://github.com/kubernetes/minikube.git  # replace path with your github fork of minikube 
cd minikube/site
hugo server # to server site locally
```

The theme is included as a Git submodule:

```bash
▶ git submodule
 2536303cad19991c673037f4f16332075141364a themes/docsy (2536303)
```

If you want to do SCSS edits and want to publish these, you need to install `PostCSS` (not needed for `hugo server`):

```bash
npm install
```
### Common Issues
```bash
Start building sites …
hugo v0.86.0+extended darwin/amd64 BuildDate=unknown
Error: Error building site: "/minikube/site/content/en/docs/contrib/releasing/binaries.md:64:1": failed to extract shortcode: template for shortcode "alert" not found
Built in 667 ms
```
This indicates the submodules are not updated. 
Please run the following command to fix.          
```  git submodule update --init --recursive ```

