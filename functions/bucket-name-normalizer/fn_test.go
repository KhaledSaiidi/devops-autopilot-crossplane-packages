package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composite"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRunFunction(t *testing.T) {
	cases := map[string]struct {
		req                 *fnv1.RunFunctionRequest
		wantBucketName      string
		wantCondition       fnv1.Status
		wantConditionReason string
		wantFatalMessage    string
		wantExistingStatus  string
	}{
		"NormalizesBucketNameFromCompositePrefix": {
			req:                 newRequest("aws", "MyBucket_Example", 63, nil),
			wantBucketName:      "mybucket-example",
			wantCondition:       fnv1.Status_STATUS_CONDITION_TRUE,
			wantConditionReason: "Success",
		},
		"UsesDesiredCompositeAsBaseAndTruncatesName": {
			req: newRequest("azure", "My___Very___Long___Bucket___Name", 12, resource.MustStructJSON(`{
				"apiVersion": "devops-autopilot-crossplane-packages.io/v1alpha1",
				"kind": "XBucket",
				"spec": {
					"parameters": {
						"namePrefix": "My___Very___Long___Bucket___Name"
					}
				},
				"status": {
					"existing": "keep-me"
				}
			}`)),
			wantBucketName:      "my-very-long",
			wantCondition:       fnv1.Status_STATUS_CONDITION_TRUE,
			wantConditionReason: "Success",
			wantExistingStatus:  "keep-me",
		},
		"FailsWhenSourceFieldIsMissing": {
			req: &fnv1.RunFunctionRequest{
				Meta: &fnv1.RequestMeta{Tag: "hello"},
				Input: resource.MustStructJSON(`{
					"apiVersion": "functions.devops-autopilot-crossplane-packages.io/v1beta1",
					"kind": "NameNormalizerInput",
					"provider": "aws",
					"sourceFieldPath": "spec.parameters.missing",
					"targetFieldPath": "status.bucketName",
					"maxLength": 63
				}`),
				Observed: &fnv1.State{
					Composite: &fnv1.Resource{
						Resource: resource.MustStructJSON(`{
							"apiVersion": "devops-autopilot-crossplane-packages.io/v1alpha1",
							"kind": "XBucket",
							"spec": {
								"parameters": {
									"namePrefix": "MyBucket_Example"
								}
							}
						}`),
					},
				},
			},
			wantCondition:       fnv1.Status_STATUS_CONDITION_FALSE,
			wantConditionReason: "InvalidInput",
			wantFatalMessage:    `cannot read source field path "spec.parameters.missing"`,
		},
		"FailsWhenNameNormalizesToEmpty": {
			req:                 newRequest("aws", "___", 63, nil),
			wantCondition:       fnv1.Status_STATUS_CONDITION_FALSE,
			wantConditionReason: "InvalidInput",
			wantFatalMessage:    "source value does not contain any valid bucket name characters",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("RunFunction() returned unexpected error: %v", err)
			}

			if got := rsp.GetMeta().GetTag(); got != "hello" {
				t.Fatalf("response tag = %q, want %q", got, "hello")
			}

			if len(rsp.GetConditions()) != 1 {
				t.Fatalf("conditions length = %d, want 1", len(rsp.GetConditions()))
			}

			condition := rsp.GetConditions()[0]
			if got := condition.GetStatus(); got != tc.wantCondition {
				t.Fatalf("condition status = %v, want %v", got, tc.wantCondition)
			}
			if got := condition.GetReason(); got != tc.wantConditionReason {
				t.Fatalf("condition reason = %q, want %q", got, tc.wantConditionReason)
			}

			if tc.wantBucketName != "" {
				xr := getCompositeFromResponse(t, rsp)
				gotBucketName, err := xr.Resource.GetString("status.bucketName")
				if err != nil {
					t.Fatalf("GetString(status.bucketName) returned error: %v", err)
				}
				if gotBucketName != tc.wantBucketName {
					t.Fatalf("bucket name = %q, want %q", gotBucketName, tc.wantBucketName)
				}

				if tc.wantExistingStatus != "" {
					gotExistingStatus, err := xr.Resource.GetString("status.existing")
					if err != nil {
						t.Fatalf("GetString(status.existing) returned error: %v", err)
					}
					if gotExistingStatus != tc.wantExistingStatus {
						t.Fatalf("existing status = %q, want %q", gotExistingStatus, tc.wantExistingStatus)
					}
				}
			}

			if tc.wantFatalMessage != "" {
				if len(rsp.GetResults()) == 0 {
					t.Fatalf("expected fatal results, got none")
				}
				lastResult := rsp.GetResults()[len(rsp.GetResults())-1]
				if got := lastResult.GetSeverity(); got != fnv1.Severity_SEVERITY_FATAL {
					t.Fatalf("fatal severity = %v, want %v", got, fnv1.Severity_SEVERITY_FATAL)
				}
				if got := lastResult.GetMessage(); got == "" || !contains(got, tc.wantFatalMessage) {
					t.Fatalf("fatal message = %q, want substring %q", got, tc.wantFatalMessage)
				}
			}
		})
	}
}

func TestNormalizeBucketName(t *testing.T) {
	cases := map[string]struct {
		provider  string
		raw       string
		maxLength int
		want      string
		wantErr   string
	}{
		"SanitizesMixedCaseAndSeparators": {
			provider:  "aws",
			raw:       " My.Bucket__Name ",
			maxLength: 63,
			want:      "my-bucket-name",
		},
		"TrimsTrailingSeparatorAfterTruncation": {
			provider:  "aws",
			raw:       "bucket--name",
			maxLength: 7,
			want:      "bucket",
		},
		"RejectsShortNormalizedName": {
			provider:  "azure",
			raw:       "A",
			maxLength: 63,
			wantErr:   "must be at least 3 characters long",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := normalizeBucketName(tc.provider, tc.raw, tc.maxLength)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("normalizeBucketName() error = nil, want %q", tc.wantErr)
				}
				if !contains(err.Error(), tc.wantErr) {
					t.Fatalf("normalizeBucketName() error = %q, want substring %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeBucketName() returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeBucketName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func newRequest(provider, namePrefix string, maxLength int, desiredComposite *structpb.Struct) *fnv1.RunFunctionRequest {
	req := &fnv1.RunFunctionRequest{
		Meta: &fnv1.RequestMeta{Tag: "hello"},
		Input: resource.MustStructJSON(fmt.Sprintf(`{
			"apiVersion": "functions.devops-autopilot-crossplane-packages.io/v1beta1",
			"kind": "NameNormalizerInput",
			"provider": %q,
			"sourceFieldPath": "spec.parameters.namePrefix",
			"targetFieldPath": "status.bucketName",
			"maxLength": %d
		}`, provider, maxLength)),
		Observed: &fnv1.State{
			Composite: &fnv1.Resource{
				Resource: resource.MustStructJSON(fmt.Sprintf(`{
					"apiVersion": "devops-autopilot-crossplane-packages.io/v1alpha1",
					"kind": "XBucket",
					"spec": {
						"parameters": {
							"namePrefix": %q
						}
					}
				}`, namePrefix)),
			},
		},
	}

	if desiredComposite != nil {
		req.Desired = &fnv1.State{
			Composite: &fnv1.Resource{
				Resource: desiredComposite,
			},
		}
	}

	return req
}

func getCompositeFromResponse(t *testing.T, rsp *fnv1.RunFunctionResponse) *resource.Composite {
	t.Helper()

	xr := &resource.Composite{Resource: composite.New()}
	if err := resource.AsObject(rsp.GetDesired().GetComposite().GetResource(), xr.Resource); err != nil {
		t.Fatalf("AsObject() returned error: %v", err)
	}

	return xr
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
