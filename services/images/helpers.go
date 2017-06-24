package images

import (
	imagesapi "github.com/containerd/containerd/api/services/images/v1"
	"github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/identifiers"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/namespaces"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func imagesToProto(images []images.Image) []imagesapi.Image {
	var imagespb []imagesapi.Image

	for _, image := range images {
		imagespb = append(imagespb, imageToProto(&image))
	}

	return imagespb
}

func imagesFromProto(imagespb []imagesapi.Image) []images.Image {
	var images []images.Image

	for _, image := range imagespb {
		images = append(images, imageFromProto(&image))
	}

	return images
}

func imageToProto(image *images.Image) imagesapi.Image {
	return imagesapi.Image{
		Name:   image.Name,
		Target: descToProto(&image.Target),
	}
}

func imageFromProto(imagepb *imagesapi.Image) images.Image {
	return images.Image{
		Name:   imagepb.Name,
		Target: descFromProto(&imagepb.Target),
	}
}

func descFromProto(desc *types.Descriptor) ocispec.Descriptor {
	return ocispec.Descriptor{
		MediaType: desc.MediaType,
		Size:      desc.Size_,
		Digest:    desc.Digest,
	}
}

func descToProto(desc *ocispec.Descriptor) types.Descriptor {
	return types.Descriptor{
		MediaType: desc.MediaType,
		Size_:     desc.Size,
		Digest:    desc.Digest,
	}
}

func rewriteGRPCError(err error) error {
	if err == nil {
		return err
	}

	switch grpc.Code(errors.Cause(err)) {
	case codes.AlreadyExists:
		return metadata.ErrExists(grpc.ErrorDesc(err))
	case codes.NotFound:
		return metadata.ErrNotFound(grpc.ErrorDesc(err))
	}

	return err
}

func mapGRPCError(err error, id string) error {
	switch {
	case metadata.IsNotFound(err):
		return grpc.Errorf(codes.NotFound, "image %v not found", id)
	case metadata.IsExists(err):
		return grpc.Errorf(codes.AlreadyExists, "image %v already exists", id)
	case namespaces.IsNamespaceRequired(err):
		return grpc.Errorf(codes.InvalidArgument, "namespace required, please set %q header", namespaces.GRPCHeader)
	case identifiers.IsInvalid(err):
		return grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	return err
}