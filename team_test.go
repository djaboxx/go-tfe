package tfe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamsList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	orgTest, orgTestCleanup := createOrganization(t, client)
	defer orgTestCleanup()

	tmTest1, tmTest1Cleanup := createTeam(t, client, orgTest)
	defer tmTest1Cleanup()
	tmTest2, tmTest2Cleanup := createTeam(t, client, orgTest)
	defer tmTest2Cleanup()

	t.Run("without list options", func(t *testing.T) {
		tms, err := client.Teams.List(ctx, orgTest.Name, TeamListOptions{})
		require.NoError(t, err)
		assert.Contains(t, tms, tmTest1)
		assert.Contains(t, tms, tmTest2)
	})

	t.Run("with list options", func(t *testing.T) {
		t.Skip("paging not supported yet in API")
		// Request a page number which is out of range. The result should
		// be successful, but return no results if the paging options are
		// properly passed along.
		tms, err := client.Teams.List(ctx, orgTest.Name, TeamListOptions{
			ListOptions: ListOptions{
				PageNumber: 999,
				PageSize:   100,
			},
		})
		require.NoError(t, err)
		assert.Empty(t, tms)
	})

	t.Run("without a valid organization", func(t *testing.T) {
		tms, err := client.Teams.List(ctx, badIdentifier, TeamListOptions{})
		assert.Nil(t, tms)
		assert.EqualError(t, err, "Invalid value for organization")
	})
}

func TestTeamsCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	orgTest, orgTestCleanup := createOrganization(t, client)
	defer orgTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := TeamCreateOptions{
			Name: String("foo"),
		}

		tm, err := client.Teams.Create(ctx, orgTest.Name, options)
		require.NoError(t, err)

		// Get a refreshed view from the API.
		refreshed, err := client.Teams.Read(ctx, tm.ID)
		require.NoError(t, err)

		for _, item := range []*Team{
			tm,
			refreshed,
		} {
			assert.NotEmpty(t, item.ID)
			assert.Equal(t, *options.Name, item.Name)
		}
	})

	t.Run("when options is missing name", func(t *testing.T) {
		tm, err := client.Teams.Create(ctx, "foo", TeamCreateOptions{})
		assert.Nil(t, tm)
		assert.EqualError(t, err, "Name is required")
	})

	t.Run("when options has an invalid name", func(t *testing.T) {
		tm, err := client.Teams.Create(ctx, "foo", TeamCreateOptions{
			Name: String(badIdentifier),
		})
		assert.Nil(t, tm)
		assert.EqualError(t, err, "Invalid value for name")
	})

	t.Run("when options has an invalid organization", func(t *testing.T) {
		tm, err := client.Teams.Create(ctx, badIdentifier, TeamCreateOptions{
			Name: String("foo"),
		})
		assert.Nil(t, tm)
		assert.EqualError(t, err, "Invalid value for organization")
	})
}

func TestTeamsRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	orgTest, orgTestCleanup := createOrganization(t, client)
	defer orgTestCleanup()

	tmTest, tmTestCleanup := createTeam(t, client, orgTest)
	defer tmTestCleanup()

	t.Run("when the team exists", func(t *testing.T) {
		tm, err := client.Teams.Read(ctx, tmTest.ID)
		require.NoError(t, err)
		assert.Equal(t, tmTest, tm)

		t.Run("permissions are properly decoded", func(t *testing.T) {
			assert.True(t, tm.Permissions.CanDestroy)
		})
	})

	t.Run("when the team does not exist", func(t *testing.T) {
		tm, err := client.Teams.Read(ctx, "nonexisting")
		assert.Nil(t, tm)
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("without a team ID", func(t *testing.T) {
		tm, err := client.Teams.Read(ctx, badIdentifier)
		assert.Nil(t, tm)
		assert.EqualError(t, err, "Invalid value for team ID")
	})
}

func TestTeamsDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	orgTest, orgTestCleanup := createOrganization(t, client)
	defer orgTestCleanup()

	tmTest, _ := createTeam(t, client, orgTest)

	t.Run("with valid options", func(t *testing.T) {
		err := client.Teams.Delete(ctx, tmTest.ID)
		require.NoError(t, err)

		// Try loading the workspace - it should fail.
		_, err = client.Teams.Read(ctx, tmTest.ID)
		assert.Equal(t, err, ErrResourceNotFound)
	})

	t.Run("without valid team ID", func(t *testing.T) {
		err := client.Teams.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "Invalid value for team ID")
	})
}
