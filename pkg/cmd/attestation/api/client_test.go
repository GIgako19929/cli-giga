package api

import (
	"testing"

	"github.com/cli/cli/v2/pkg/cmd/attestation/io"

	"github.com/stretchr/testify/require"
)

const (
	testRepo   = "github/example"
	testOwner  = "github"
	testDigest = "sha256:12313213"
)

func NewClientWithMockGHClient(hasNextPage bool) Client {
	fetcher := mockDataGenerator{
		NumAttestations: 5,
	}
	l := io.NewTestHandler()

	if hasNextPage {
		return &LiveClient{
			api: mockAPIClient{
				OnRESTWithNext: fetcher.OnRESTSuccessWithNextPage,
			},
			logger: l,
		}
	}

	return &LiveClient{
		api: mockAPIClient{
			OnRESTWithNext: fetcher.OnRESTSuccess,
		},
		logger: l,
	}
}

func TestGetURL(t *testing.T) {
	c := LiveClient{}

	testData := []struct {
		repo     string
		digest   string
		expected string
	}{
		{repo: "/github/example/", digest: "sha256:12313213", expected: "repos/github/example/attestations/sha256:12313213"},
		{repo: "/github/example", digest: "sha256:12313213", expected: "repos/github/example/attestations/sha256:12313213"},
	}

	for _, data := range testData {
		s := c.BuildRepoAndDigestURL(data.repo, data.digest)
		require.Equal(t, data.expected, s)
	}
}

func TestGetByDigest(t *testing.T) {
	c := NewClientWithMockGHClient(false)
	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, DefaultLimit)
	require.NoError(t, err)

	require.Equal(t, 5, len(attestations))
	bundle := (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")

	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, DefaultLimit)
	require.NoError(t, err)

	require.Equal(t, 5, len(attestations))
	bundle = (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
}

func TestGetByDigestGreaterThanLimit(t *testing.T) {
	c := NewClientWithMockGHClient(false)

	limit := 3
	// The method should return five results when the limit is not set
	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, limit)
	require.NoError(t, err)

	require.Equal(t, 3, len(attestations))
	bundle := (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")

	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, limit)
	require.NoError(t, err)

	require.Equal(t, len(attestations), limit)
	bundle = (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
}

func TestGetByDigestWithNextPage(t *testing.T) {
	c := NewClientWithMockGHClient(true)
	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, DefaultLimit)
	require.NoError(t, err)

	require.Equal(t, len(attestations), 10)
	bundle := (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")

	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, DefaultLimit)
	require.NoError(t, err)

	require.Equal(t, len(attestations), 10)
	bundle = (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
}

func TestGetByDigestGreaterThanLimitWithNextPage(t *testing.T) {
	c := NewClientWithMockGHClient(true)

	limit := 7
	// The method should return five results when the limit is not set
	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, limit)
	require.NoError(t, err)

	require.Equal(t, len(attestations), limit)
	bundle := (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")

	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, limit)
	require.NoError(t, err)

	require.Equal(t, len(attestations), limit)
	bundle = (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
}

func TestGetByDigest_NoAttestationsFound(t *testing.T) {
	fetcher := mockDataGenerator{
		NumAttestations: 5,
	}

	c := LiveClient{
		api: mockAPIClient{
			OnRESTWithNext: fetcher.OnRESTWithNextNoAttestations,
		},
		logger: io.NewTestHandler(),
	}

	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, DefaultLimit)
	require.Error(t, err)
	require.IsType(t, ErrNoAttestations{}, err)
	require.Nil(t, attestations)

	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, DefaultLimit)
	require.Error(t, err)
	require.IsType(t, ErrNoAttestations{}, err)
	require.Nil(t, attestations)
}

func TestGetByDigest_Error(t *testing.T) {
	fetcher := mockDataGenerator{
		NumAttestations: 5,
	}

	c := LiveClient{
		api: mockAPIClient{
			OnRESTWithNext: fetcher.OnRESTWithNextError,
		},
		logger: io.NewTestHandler(),
	}

	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, DefaultLimit)
	require.Error(t, err)
	require.Nil(t, attestations)

	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, DefaultLimit)
	require.Error(t, err)
	require.Nil(t, attestations)
}

func TestGetTrustDomain(t *testing.T) {
	fetcher := mockMetaGenerator{
		TrustDomain: "foo",
	}

	t.Run("with returned trust domain", func(t *testing.T) {
		c := LiveClient{
			api: mockAPIClient{
				OnREST: fetcher.OnREST,
			},
			logger: io.NewTestHandler(),
		}
		td, err := c.GetTrustDomain()
		require.Nil(t, err)
		require.Equal(t, "foo", td)

	})

	t.Run("with error", func(t *testing.T) {
		c := LiveClient{
			api: mockAPIClient{
				OnREST: fetcher.OnRESTError,
			},
			logger: io.NewTestHandler(),
		}
		td, err := c.GetTrustDomain()
		require.Equal(t, "", td)
		require.ErrorContains(t, err, "test error")
	})

}

func TestGetAttestationsRetries(t *testing.T) {
	getAttestationRetryInterval = 0

	fetcher := mockDataGenerator{
		NumAttestations: 5,
	}

	c := &LiveClient{
		api: mockAPIClient{
			OnRESTWithNext: fetcher.FlakyOnRESTSuccessWithNextPageHandler(),
		},
		logger: io.NewTestHandler(),
	}

	attestations, err := c.GetByRepoAndDigest(testRepo, testDigest, DefaultLimit)
	require.NoError(t, err)

	// assert the error path was executed; because this is a paged
	// request, it should have errored twice
	fetcher.AssertNumberOfCalls(t, "FlakyOnRESTSuccessWithNextPage:error", 2)

	// but we still successfully got the right data
	require.Equal(t, len(attestations), 10)
	bundle := (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")

	// same test as above, but for GetByOwnerAndDigest:
	attestations, err = c.GetByOwnerAndDigest(testOwner, testDigest, DefaultLimit)
	require.NoError(t, err)

	// because we haven't reset the mock, we have added 2 more failed requests
	fetcher.AssertNumberOfCalls(t, "FlakyOnRESTSuccessWithNextPage:error", 4)

	require.Equal(t, len(attestations), 10)
	bundle = (attestations)[0].Bundle
	require.Equal(t, bundle.GetMediaType(), "application/vnd.dev.sigstore.bundle.v0.3+json")
}

// test total retries
func TestGetAttestationsMaxRetries(t *testing.T) {
	getAttestationRetryInterval = 0

	fetcher := mockDataGenerator{
		NumAttestations: 5,
	}

	c := &LiveClient{
		api: mockAPIClient{
			OnRESTWithNext: fetcher.OnREST500ErrorHandler(),
		},
		logger: io.NewTestHandler(),
	}

	_, err := c.GetByRepoAndDigest(testRepo, testDigest, DefaultLimit)
	require.Error(t, err)

	fetcher.AssertNumberOfCalls(t, "OnREST500Error", 4)
}
