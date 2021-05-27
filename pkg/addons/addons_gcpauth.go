/*
Copyright 2021 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addons

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	credentialsPath = "/var/lib/minikube/google_application_credentials.json"
	projectPath     = "/var/lib/minikube/google_cloud_project"
	secretName      = "gcp-auth"
	namespaceName   = "gcp-auth"
)

// enableOrDisableGCPAuth enables or disables the gcp-auth addon depending on the val parameter
func enableOrDisableGCPAuth(cfg *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddonGCPAuth(cfg)
	}
	return disableAddonGCPAuth(cfg)
}

func enableAddonGCPAuth(cfg *config.ClusterConfig) error {
	if !Force && detect.IsOnGCE() {
		exit.Message(reason.InternalCredsNotNeeded, "It seems that you are running in GCE, which means authentication should work without the GCP Auth addon. If you would still like to authenticate using a credentials file, use the --force flag.")
	}

	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Grab credentials from where GCP would normally look
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil || creds.JSON == nil {
		exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
	}

	// Actually copy the creds over
	f := assets.NewMemoryAssetTarget(creds.JSON, credentialsPath, "0444")

	err = r.Copy(f)
	if err != nil {
		return err
	}

	// Create a registry secret in every namespace we can find
	err = createPullSecret(cfg, creds)
	if err != nil {
		return errors.Wrap(err, "pull secret")
	}

	// First check if the project env var is explicitly set
	projectEnv := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectEnv != "" {
		f := assets.NewMemoryAssetTarget([]byte(projectEnv), projectPath, "0444")
		return r.Copy(f)
	}

	// We're currently assuming gcloud is installed and in the user's path
	proj, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err == nil && len(proj) > 0 {
		f := assets.NewMemoryAssetTarget(bytes.TrimSpace(proj), projectPath, "0444")
		return r.Copy(f)
	}

	out.WarningT("Could not determine a Google Cloud project, which might be ok.")
	out.Styled(style.Tip, `To set your Google Cloud project,  run:

		gcloud config set project <project name>

or set the GOOGLE_CLOUD_PROJECT environment variable.`)

	// Copy an empty file in to avoid errors about missing files
	emptyFile := assets.NewMemoryAssetTarget([]byte{}, projectPath, "0444")
	return r.Copy(emptyFile)

}

func createPullSecret(cc *config.ClusterConfig, creds *google.Credentials) error {
	client, err := service.K8s.GetCoreClient(cc.Name)
	if err != nil {
		return err
	}

	namespaces, err := client.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	token, err := creds.TokenSource.Token()
	// Only try to add secret if Token was found
	if err == nil {
		data := map[string][]byte{
			".dockercfg": []byte(fmt.Sprintf(`{"https://gcr.io":{"username":"oauth2accesstoken","password":"%s","email":"none"}, "https://us-docker.pkg.dev":{"username":"oauth2accesstoken","password":"%s","email":"none"}}`, token.AccessToken, token.AccessToken)),
		}

		for _, n := range namespaces.Items {
			secrets := client.Secrets(n.Name)

			exists := false
			secList, err := secrets.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, s := range secList.Items {
				if s.Name == secretName {
					exists = true
					break
				}
			}

			if !exists {
				secretObj := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: secretName,
					},
					Data: data,
					Type: "kubernetes.io/dockercfg",
				}

				_, err = secrets.Create(context.TODO(), secretObj, metav1.CreateOptions{})
				if err != nil {
					return err
				}
			}

			// Now patch the secret into all the service accounts we can find
			serviceaccounts := client.ServiceAccounts(n.Name)
			salist, err := serviceaccounts.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			// Let's make sure we at least find the default service account
			for len(salist.Items) == 0 {
				salist, err = serviceaccounts.List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					return err
				}
				time.Sleep(1 * time.Second)
			}

			ips := corev1.LocalObjectReference{Name: secretName}
			for _, sa := range salist.Items {
				sa.ImagePullSecrets = append(sa.ImagePullSecrets, ips)
				_, err := serviceaccounts.Update(context.TODO(), &sa, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}

func refreshExistingPods(cc *config.ClusterConfig) error {
	client, err := service.K8s.GetCoreClient(cc.Name)
	if err != nil {
		return err
	}

	namespaces, err := client.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, n := range namespaces.Items {
		// Ignore kube-system and gcp-auth namespaces
		if n.Name == metav1.NamespaceSystem || n.Name == namespaceName {
			continue
		}

		pods := client.Pods(n.Name)
		podList, err := pods.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, p := range podList.Items {
			// Skip pods we're explicitly told to skip
			if _, ok := p.Labels["gcp-auth-skip-secret"]; ok {
				continue
			}

			// Recreating the pod should pickup the necessary changes
			err := pods.Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			p.ResourceVersion = ""

			_, err = pods.Get(context.TODO(), p.Name, metav1.GetOptions{})

			for err == nil {
				time.Sleep(time.Second)
				_, err = pods.Get(context.TODO(), p.Name, metav1.GetOptions{})
			}

			_, err = pods.Create(context.TODO(), &p, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func disableAddonGCPAuth(cfg *config.ClusterConfig) error {
	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Clean up the files generated when enabling the addon
	creds := assets.NewMemoryAssetTarget([]byte{}, credentialsPath, "0444")
	err := r.Remove(creds)
	if err != nil {
		return err
	}

	project := assets.NewMemoryAssetTarget([]byte{}, projectPath, "0444")
	err = r.Remove(project)
	if err != nil {
		return err
	}

	client, err := service.K8s.GetCoreClient(cfg.Name)
	if err != nil {
		return err
	}

	namespaces, err := client.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// No need to check for an error here, if the secret doesn't exist, no harm done.
	for _, n := range namespaces.Items {
		secrets := client.Secrets(n.Name)
		err := secrets.Delete(context.TODO(), secretName, metav1.DeleteOptions{})
		if err != nil {
			klog.Infof("error deleting secret: %v", err)
		}
	}

	return nil
}

func verifyGCPAuthAddon(cc *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	err = verifyAddonStatusInternal(cc, name, val, "gcp-auth")
	if err != nil {
		return err
	}

	if Refresh {
		err = refreshExistingPods(cc)
		if err != nil {
			return err
		}
	}

	if enable && err == nil {
		out.Styled(style.Notice, "Your GCP credentials will now be mounted into every pod created in the {{.name}} cluster.", out.V{"name": cc.Name})
		out.Styled(style.Notice, "If you don't want your credentials mounted into a specific pod, add a label with the `gcp-auth-skip-secret` key to your pod configuration.")
		if !Refresh {
			out.Styled(style.Notice, "If you want existing pods to be mounted with credentials, either recreate them or rerun addons enable with --refresh.")
		}
	}

	return err
}
