package api

import (
	"context"
	"fmt"

	"github.com/superfly/graphql"
)

type OrganizationType string

const (
	OrganizationTypePersonal OrganizationType = "PERSONAL"
	OrganizationTypeShared   OrganizationType = "SHARED"
)

type organizationFilter struct {
	admin bool
}

func (f *organizationFilter) apply(req *graphql.Request) {
	req.Var("admin", f.admin)
}

type OrganizationFilter func(*organizationFilter)

var AdminOnly OrganizationFilter = func(f *organizationFilter) { f.admin = true }

func (client *Client) GetOrganizations(ctx context.Context, filters ...OrganizationFilter) ([]Organization, error) {
	q := `
		query($admin: Boolean!) {
			organizations(admin: $admin) {
				nodes {
					id
					slug
					name
					type
					paidPlan
				}
			}
		}
	`

	filter := new(organizationFilter)
	for _, f := range filters {
		f(filter)
	}

	req := client.NewRequest(q)
	filter.apply(req)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return data.Organizations.Nodes, nil
}

func (client *Client) GetOrganizationBySlug(ctx context.Context, slug string) (*Organization, error) {
	q := `
		query($slug: String!) {
			organization(slug: $slug) {
				id
				internalNumericId
				slug
				name
				type
                limitedAccessTokens {
					nodes {
					    id
					    name
					    expiresAt
					}
                }
			}
		}
	`

	req := client.NewRequest(q)

	req.Var("slug", slug)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return data.Organization, nil
}

func (client *Client) GetCurrentOrganizations(ctx context.Context) (Organization, []Organization, error) {
	query := `
	query {
		personalOrganization {
		  id
		  slug
		  name
		  type
		  viewerRole
		}
		organizations {
		  nodes {
			id
			slug
			name
			type
			viewerRole
		  }
		}
	  }
	`

	req := client.NewRequest(query)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return Organization{}, nil, err
	}
	return data.PersonalOrganization, data.Organizations.Nodes, nil
}

func (client *Client) GetDetailedOrganizationBySlug(ctx context.Context, slug string) (*OrganizationDetails, error) {
	query := `query($slug: String!) {
		organizationdetails: organization(slug: $slug) {
		  id
		  slug
		  name
		  type
		  viewerRole
		  internalNumericId
			remoteBuilderImage
			remoteBuilderApp {
				name
			}
			members {
				edges {
					cursor
					node {
						id
						name
						email
					}
					joinedAt
					role
				}
		  }
		}
	  }
	`

	req := client.NewRequest(query)
	req.Var("slug", slug)

	data, err := client.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.OrganizationDetails, nil
}

func (c *Client) CreateOrganization(ctx context.Context, organizationname string) (*Organization, error) {
	query := `
		mutation($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
			    organization {
					id
					name
					slug
					type
					viewerRole
				  }
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]string{
		"name": organizationname,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.CreateOrganization.Organization, nil
}

func (c *Client) CreateOrganizationWithAppsV2DefaultOn(ctx context.Context, organizationname string) (*Organization, error) {
	query := `
		mutation($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
			    organization {
					id
					name
					slug
					type
					viewerRole
				  }
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]interface{}{
		"name":            organizationname,
		"appsV2DefaultOn": true,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.CreateOrganization.Organization, nil
}

func (c *Client) DeleteOrganization(ctx context.Context, id string) (deletedid string, err error) {
	query := `
	mutation($input: DeleteOrganizationInput!) {
		deleteOrganization(input: $input) {
		  clientMutationId
		  deletedOrganizationId
		  }
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]string{
		"organizationId": id,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return "", err
	}

	return data.DeleteOrganization.DeletedOrganizationId, nil
}

func (c *Client) CreateOrganizationInvite(ctx context.Context, id, email string) (*Invitation, error) {
	query := `
	mutation($input: CreateOrganizationInvitationInput!){
		createOrganizationInvitation(input: $input){
			invitation {
				id
				email
				createdAt
				redeemed
				organization {
			  		slug
				}
		  }
		}
	  }
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]string{
		"organizationId": id,
		"email":          email,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.CreateOrganizationInvitation.Invitation, nil
}

func (c *Client) DeleteOrganizationMembership(ctx context.Context, orgId, userId string) (string, string, error) {
	query := `
	mutation($input: DeleteOrganizationMembershipInput!){
		deleteOrganizationMembership(input: $input){
		organization{
		  slug
		}
		user{
		  name
		  email
		}
	  }
	}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]string{
		"userId":         userId,
		"organizationId": orgId,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return "", "", err
	}

	return data.DeleteOrganizationMembership.Organization.Name, data.DeleteOrganizationMembership.User.Email, nil
}

func (c *Client) UpdateRemoteBuilder(ctx context.Context, orgName string, image string) (*Organization, error) {
	org, err := c.GetOrganizationBySlug(ctx, orgName)
	if err != nil {
		return nil, err
	}

	query := `
		mutation($input: UpdateRemoteBuilderInput!) {
			updateRemoteBuilder(input: $input) {
			    organization {
						remoteBuilderImage
					}
			}
		}
	`

	req := c.NewRequest(query)

	req.Var("input", map[string]string{
		"organizationId": org.ID,
		"image":          image,
	})

	data, err := c.RunWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	return &data.UpdateRemoteBuilder.Organization, nil
}

const appsV2DefaultOnSettingsKey = "apps_v2_default_on"

func (c *Client) GetAppsV2DefaultOnForOrg(ctx context.Context, orgSlug string) (bool, error) {
	query := `
	query($slug: String!) {
		organization(slug: $slug) {
			settings
		}
	}
	`
	req := c.NewRequest(query)
	req.Var("slug", orgSlug)

	resp, err := c.RunWithContext(ctx, req)
	if err != nil {
		return false, err
	}

	if val, present := resp.Organization.Settings[appsV2DefaultOnSettingsKey]; !present {
		return false, nil
	} else if appsV2DefaultOn, ok := val.(bool); !ok {
		return false, fmt.Errorf("failed to convert '%v' to boolean value for %s org setting", val, appsV2DefaultOnSettingsKey)
	} else {
		return appsV2DefaultOn, nil
	}
}
