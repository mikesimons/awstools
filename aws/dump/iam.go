package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

func IAMListUsersAndAccessKeys(session *Session) *FetchResult {
	client := iam.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListUsersPages(&iam.ListUsersInput{},
		func(page *iam.ListUsersOutput, lastPage bool) bool {
			for _, user := range page.Users {
				resource, err := NewResource(*user.Arn)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)

				keysResult := IAMListAccessKeys(session, *user.UserName)
				if keysResult.Error != nil {
					result.Error = keysResult.Error
					return false
				}
				result.Resources = append(result.Resources, keysResult.Resources...)
			}

			return true
		})

	return result
}

func IAMListGroups(session *Session) *FetchResult {
	client := iam.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListGroupsPages(&iam.ListGroupsInput{},
		func(page *iam.ListGroupsOutput, lastPage bool) bool {
			for _, group := range page.Groups {

				resource, err := NewResource(*group.Arn)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}

func IAMListRoles(session *Session) *FetchResult {
	client := iam.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListRolesPages(&iam.ListRolesInput{},
		func(page *iam.ListRolesOutput, lastPage bool) bool {
			for _, role := range page.Roles {
				resource, err := NewResource(*role.Arn)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}

func IAMListPolicies(session *Session) *FetchResult {
	client := iam.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListPoliciesPages(&iam.ListPoliciesInput{Scope: aws.String("Local")},
		func(page *iam.ListPoliciesOutput, lastPage bool) bool {
			for _, policy := range page.Policies {
				resource, err := NewResource(*policy.Arn)
				if err != nil {
					result.Error = err
					return false
				}
				result.Resources = append(result.Resources, *resource)
			}

			return true
		})

	return result
}

func IAMListAccessKeys(session *Session, username string) *FetchResult {
	client := iam.New(session.Session, session.Config)

	result := &FetchResult{}
	result.Error = client.ListAccessKeysPages(&iam.ListAccessKeysInput{
		UserName: aws.String(username),
	},
		func(page *iam.ListAccessKeysOutput, lastPage bool) bool {
			for _, accessKey := range page.AccessKeyMetadata {
				result.Resources = append(result.Resources, Resource{
					ID:        *accessKey.AccessKeyId,
					AccountID: session.AccountID,
					Service:   "iam",
					Type:      "access-key",
					Metadata: map[string]string{
						"UserName": *accessKey.UserName,
					},
				})
			}

			return true
		})

	return result
}