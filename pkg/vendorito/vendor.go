package vendorito

import (
	"net/url"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
)

// ParseDockerURL parses a docker URI to retrieve its Docker image reference and URL options
func ParseDockerURL(img string) (types.ImageReference, *url.URL, error) {
	uri, err := url.Parse("docker://" + strings.TrimLeft(img, "/"))
	if err != nil {
		return nil, nil, err
	}

	// If host is not specified, default to Docker Hub
	if strings.IndexRune(uri.Host, '.') == -1 {
		uri.Path = uri.Host + uri.Path
		uri.Host = "registry.hub.docker.com"
	}

	// Strip down some info for the reference URL
	refUri := *uri
	refUri.User = nil
	refUri.Scheme = ""
	refImage, err := docker.Transport.ParseReference(refUri.String())

	return refImage, uri, err
}
