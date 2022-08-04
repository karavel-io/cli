package action

import (
	"context"
	"fmt"

	"github.com/karavel-io/cli/internal/kubeclient"
	"github.com/karavel-io/cli/internal/utils"
	"github.com/karavel-io/cli/pkg/logger"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type BootstrapParams struct {
	KustomizeDir string
	KubeConfig   string
}

func Bootstrap(ctx context.Context, params BootstrapParams) error {
	options := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(options)
	rendered, err := k.Run(filesys.MakeFsOnDisk(), params.KustomizeDir)
	if err != nil {
		return fmt.Errorf("kustomize render failed: %w", err)
	}

	// Put resources in two "phases", CRD/Namespaces and everything else
	prerequisites := resmap.New()
	resources := resmap.New()
	for _, res := range rendered.Resources() {
		switch res.GetKind() {
		case "Namespace", "CustomResourceDefinition":
			_ = prerequisites.Append(res)
		default:
			_ = resources.Append(res)
		}
	}

	client, err := kubeclient.NewClientFromConfig(params.KubeConfig)
	if err != nil {
		return fmt.Errorf("could not initialize kubernetes client: %w", err)
	}

	batch := utils.NewBatchJob()
	for _, prerequisite := range prerequisites.Resources() {
		batch.Add(utils.BindParam(func(res *resource.Resource) error {
			return applyResource(ctx, client, res, false)
		}, prerequisite))
	}

	if err := batch.Run(); err != nil {
		return err
	}

	// Wait some time to make sure cluster is up-to-date
	if err = client.FetchResources(); err != nil {
		return fmt.Errorf("could not get resources from cluster: %w", err)
	}

	batch = utils.NewBatchJob()
	for _, toApply := range resources.Resources() {
		batch.Add(utils.BindParam(func(res *resource.Resource) error {
			return applyResource(ctx, client, res, true)
		}, toApply))
	}

	if err := batch.Run(); err != nil {
		return err
	}

	return nil
}

func applyResource(ctx context.Context, client *kubeclient.KubeClient, resource *resource.Resource, onlySupported bool) error {
	log := logger.FromContext(ctx)

	resourceName := fmt.Sprintf("%s/%s", resource.GetKind(), resource.GetName())
	if onlySupported && !client.IsResourceSupported(resource) {
		log.Infof("skipping %s (resource not supported)", resourceName)
		return nil
	}

	if err := client.ApplyResource(ctx, resource); err != nil {
		return fmt.Errorf("could not apply %s: %w", resourceName, err)
	}

	log.Infof("updated %s", resourceName)
	return nil
}
