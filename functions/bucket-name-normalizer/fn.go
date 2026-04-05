package main

import (
	"context"
	"strings"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/khaledsaiidi/devops-autopilot-crossplane-packages/functions/function-bucket-name-normalizer/input/v1beta1"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.NameNormalizerInput{}
	if err := request.GetInput(req, in); err != nil {
		err = errors.Wrapf(err, "cannot get function input from %T", req)
		f.fail(rsp, "InvalidInput", err)
		return rsp, nil
	}

	xr, err := getBaseComposite(req)
	if err != nil {
		err = errors.Wrap(err, "cannot get composite resource")
		f.fail(rsp, "InvalidComposite", err)
		return rsp, nil
	}

	sourceValue, err := xr.Resource.GetString(in.SourceFieldPath)
	if err != nil {
		err = errors.Wrapf(err, "cannot read source field path %q", in.SourceFieldPath)
		f.fail(rsp, "InvalidInput", err)
		return rsp, nil
	}

	bucketName, err := normalizeBucketName(in.Provider, sourceValue, in.MaxLength)
	if err != nil {
		err = errors.Wrap(err, "cannot normalize bucket name")
		f.fail(rsp, "InvalidInput", err)
		return rsp, nil
	}

	if err := xr.Resource.SetString(in.TargetFieldPath, bucketName); err != nil {
		err = errors.Wrapf(err, "cannot write target field path %q", in.TargetFieldPath)
		f.fail(rsp, "InternalError", err)
		return rsp, nil
	}

	if err := response.SetDesiredCompositeResource(rsp, xr); err != nil {
		err = errors.Wrap(err, "cannot set desired composite resource")
		f.fail(rsp, "InternalError", err)
		return rsp, nil
	}

	response.Normalf(rsp, "Normalized %q to bucket name %q.", sourceValue, bucketName)
	f.log.Info("Normalized bucket name",
		"provider", strings.ToLower(strings.TrimSpace(in.Provider)),
		"sourceFieldPath", in.SourceFieldPath,
		"targetFieldPath", in.TargetFieldPath,
		"bucketName", bucketName)
	response.ConditionTrue(rsp, "FunctionSuccess", "Success").TargetCompositeAndClaim()

	return rsp, nil
}

func (f *Function) fail(rsp *fnv1.RunFunctionResponse, reason string, err error) {
	f.log.Info("Function failed", "reason", reason, "error", err.Error())
	response.ConditionFalse(rsp, "FunctionSuccess", reason).
		WithMessage(err.Error()).
		TargetCompositeAndClaim()
	response.Fatal(rsp, err)
}

func getBaseComposite(req *fnv1.RunFunctionRequest) (*resource.Composite, error) {
	if req.GetDesired().GetComposite().GetResource() != nil {
		return request.GetDesiredCompositeResource(req)
	}

	return request.GetObservedCompositeResource(req)
}

func normalizeBucketName(provider, raw string, maxLength int) (string, error) {
	if maxLength < 1 {
		return "", errors.New("maxLength must be greater than zero")
	}

	normalized := sanitizeName(raw)
	if normalized == "" {
		return "", errors.New("source value does not contain any valid bucket name characters")
	}

	if len(normalized) > maxLength {
		normalized = strings.Trim(normalized[:maxLength], "-")
	}

	if normalized == "" {
		return "", errors.New("normalized bucket name is empty after applying maxLength")
	}

	minLength := 3
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "aws", "azure":
		// Current bucket packages use AWS and Azure, and both require at least 3
		// characters for the names this function produces.
	default:
		// Fall back to the generic bucket constraints for any future provider.
	}

	if len(normalized) < minLength {
		return "", errors.Errorf("normalized bucket name %q must be at least %d characters long", normalized, minLength)
	}

	return normalized, nil
}

func sanitizeName(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(raw))
	lastWasSeparator := false

	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastWasSeparator = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastWasSeparator = false
		default:
			if b.Len() == 0 || lastWasSeparator {
				continue
			}
			b.WriteByte('-')
			lastWasSeparator = true
		}
	}

	return strings.Trim(b.String(), "-")
}
