package tfe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyChecksList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	orgTest, orgTestCleanup := createOrganization(t, client)
	defer orgTestCleanup()

	createUploadedPolicy(t, client, true, orgTest)
	createUploadedPolicy(t, client, true, orgTest)

	wTest, _ := createWorkspace(t, client, orgTest)
	rTest, _ := createPlannedRun(t, client, wTest)

	t.Run("without list options", func(t *testing.T) {
		pcl, err := client.PolicyChecks.List(ctx, rTest.ID, PolicyCheckListOptions{})
		require.NoError(t, err)
		require.Equal(t, 1, len(pcl.Items))

		t.Run("pagination is properly decoded", func(t *testing.T) {
			t.Skip("paging not supported yet in API")
			assert.Equal(t, 1, pcl.CurrentPage)
			assert.Equal(t, 1, pcl.TotalCount)
		})

		t.Run("permissions are properly decoded", func(t *testing.T) {
			assert.NotEmpty(t, pcl.Items[0].Permissions)
		})

		t.Run("result is properly decoded", func(t *testing.T) {
			require.NotEmpty(t, pcl.Items[0].Result)
			assert.Equal(t, 2, pcl.Items[0].Result.Passed)
		})

		t.Run("timestamps are properly decoded", func(t *testing.T) {
			assert.NotEmpty(t, pcl.Items[0].StatusTimestamps)
		})
	})

	t.Run("with list options", func(t *testing.T) {
		t.Skip("paging not supported yet in API")
		// Request a page number which is out of range. The result should
		// be successful, but return no results if the paging options are
		// properly passed along.
		pcl, err := client.PolicyChecks.List(ctx, rTest.ID, PolicyCheckListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
		})
		require.NoError(t, err)
		assert.Empty(t, pcl.Items)
		assert.Equal(t, 999, pcl.CurrentPage)
		assert.Equal(t, 1, pcl.TotalCount)
	})

	t.Run("without a valid run ID", func(t *testing.T) {
		pcl, err := client.PolicyChecks.List(ctx, badIdentifier, PolicyCheckListOptions{})
		assert.Nil(t, pcl)
		assert.EqualError(t, err, "Invalid value for run ID")
	})
}

func TestPolicyChecksOverride(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	orgTest, orgTestCleanup := createOrganization(t, client)
	defer orgTestCleanup()

	t.Run("when the policy failed", func(t *testing.T) {
		_, pTestCleanup := createUploadedPolicy(t, client, false, orgTest)
		defer pTestCleanup()

		wTest, _ := createWorkspace(t, client, orgTest)
		rTest, _ := createPlannedRun(t, client, wTest)

		pcl, err := client.PolicyChecks.List(ctx, rTest.ID, PolicyCheckListOptions{})
		require.NoError(t, err)
		require.Equal(t, 1, len(pcl.Items))
		require.Equal(t, PolicySoftFailed, pcl.Items[0].Status)

		pc, err := client.PolicyChecks.Override(ctx, pcl.Items[0].ID)
		require.NoError(t, err)

		assert.NotEmpty(t, pc.Result)
		assert.Equal(t, PolicyOverridden, pc.Status)
	})

	t.Run("when the policy passed", func(t *testing.T) {
		_, pTestCleanup := createUploadedPolicy(t, client, true, orgTest)
		defer pTestCleanup()

		wTest, _ := createWorkspace(t, client, orgTest)
		rTest, _ := createPlannedRun(t, client, wTest)

		pcl, err := client.PolicyChecks.List(ctx, rTest.ID, PolicyCheckListOptions{})
		require.NoError(t, err)
		require.Equal(t, 1, len(pcl.Items))
		require.Equal(t, PolicyPasses, pcl.Items[0].Status)

		_, err = client.PolicyChecks.Override(ctx, pcl.Items[0].ID)
		assert.Error(t, err)
	})

	t.Run("without a valid policy check ID", func(t *testing.T) {
		p, err := client.PolicyChecks.Override(ctx, badIdentifier)
		assert.Nil(t, p)
		assert.EqualError(t, err, "Invalid value for policy check ID")
	})
}
