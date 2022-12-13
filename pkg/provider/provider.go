package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/goblain/go-retry"
	"github.com/hashicorp/go-tfe"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func retryer() *retry.RetryLogic {
	rl, _ := retry.NewRetryLogic(retry.WithExponentialBackoff(time.Second, time.Second*10, 1.2))
	return rl
}

func PublishPrivateProvider(ctx context.Context, tfc *tfe.Client, org, namespace, keyID, path string) error {
	providerFilename, providerName, providerVersion, providerOS, providerArch := extractMetadataFromPath(path)
	vid := tfe.RegistryProviderVersionID{
		RegistryProviderID: tfe.RegistryProviderID{
			OrganizationName: org,
			RegistryName:     tfe.PrivateRegistry,
			Namespace:        namespace,
			Name:             providerName,
		},
		Version: providerVersion,
	}
	vi, err := retryer().ExecuteFuncI(func() (interface{}, error) {
		v, err := getOrCreateVersion(ctx, tfc, vid, keyID)
		if err != nil {
			return nil, fmt.Errorf("version: %w", err)
		}
		if !strings.HasSuffix(path, "SHA256SUMS") && !strings.HasSuffix(path, "SHA256SUMS.sig") {
			if !v.ShasumsUploaded || !v.ShasumsSigUploaded {
				return nil, fmt.Errorf("waiting for shasums")
			}
		}
		return v, nil
	})
	if err != nil {
		return err
	}
	v := vi.(*tfe.RegistryProviderVersion)
	var url string
	switch {
	case strings.HasSuffix(path, "SHA256SUMS"):
		url = v.Links["shasums-upload"].(string)
		return upload(ctx, url, path)
	case strings.HasSuffix(path, "SHA256SUMS.sig"):
		url = v.Links["shasums-sig-upload"].(string)
		return upload(ctx, url, path)
	case strings.HasSuffix(path, ".zip"):
		return uploadPlatform(ctx, tfc, vid, path, providerFilename, providerOS, providerArch)
	}
	return fmt.Errorf("unsupported artifact %s", path)
}

func extractMetadataFromPath(filePath string) (file, provider, version, os, arch string) {
	re := regexp.MustCompile("terraform-provider-([a-zA-Z0-9]+)_([a-zA-Z0-9.]+)(-[a-zA-Z]+)?_?([a-zA-Z0-9]+)?_?([a-zA-Z0-9]+)?.?[a-zA-Z]*")
	found := re.FindAllStringSubmatch(filePath, 1)
	file = found[0][0]
	provider = found[0][1]
	version = found[0][2] + found[0][3]
	os = found[0][4]
	arch = found[0][5]
	return
}

func uploadPlatform(ctx context.Context, tfc *tfe.Client, vid tfe.RegistryProviderVersionID, path, providerFilename, providerOS, providerArch string) error {
	opts := tfe.RegistryProviderPlatformCreateOptions{}
	opts.Filename = providerFilename
	opts.OS = providerOS
	opts.Arch = providerArch
	hasher := sha256.New()
	s, err := os.ReadFile(path)
	hasher.Write(s)
	opts.Shasum = hex.EncodeToString(hasher.Sum(nil))
	platform, err := tfc.RegistryProviderPlatforms.Create(ctx, vid, opts)
	if err != nil {
		return err
	}
	if platform.ProviderBinaryUploaded {
		return fmt.Errorf("binary already uploaded")
	}
	upload(ctx, platform.Links["provider-binary-upload"].(string), path)
	return nil
}

func upload(_ context.Context, url, path string) error {
	fh, err := os.Open(path)
	hc, _ := http.NewRequest(http.MethodPut, url, fh)
	resp, err := http.DefaultClient.Do(hc)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received %d instead of 200: %s", resp.StatusCode, string(body))
	}
	return nil
}

func getOrCreateVersion(ctx context.Context, tfc *tfe.Client, vid tfe.RegistryProviderVersionID, keyID string) (*tfe.RegistryProviderVersion, error) {
	for {
		opts := &tfe.RegistryProviderVersionListOptions{}
		vl, err := tfc.RegistryProviderVersions.List(ctx, vid.RegistryProviderID, opts)
		if err != nil {
			return nil, err
		}
		for _, v := range vl.Items {
			if v.Version == vid.Version {
				return v, nil
			}
		}
		if vl.CurrentPage == vl.TotalPages {
			break
		}
		opts.PageNumber = vl.NextPage
	}

	v, err := tfc.RegistryProviderVersions.Create(ctx, vid.RegistryProviderID, tfe.RegistryProviderVersionCreateOptions{
		Version: vid.Version,
		KeyID:   keyID,
	})
	if err != nil {
		return nil, err
	}

	return v, nil
}
