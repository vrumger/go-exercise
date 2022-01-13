package cmd

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var token string

var githubCmd = &cobra.Command{
	Use: "github",
	Run: func(cmd *cobra.Command, args []string) {
		if token == "" {
			fmt.Println("Missing token")
			return
		}

		ctx := context.Background()
		tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tokenClient := oauth2.NewClient(ctx, tokenSource)

		client := github.NewClient(tokenClient)

		orgs, _, err := client.Organizations.List(context.Background(), "", nil)
		if err != nil {
			panic(err)
		}

		for _, org := range orgs {
			tfaDisabledMembers, _, _ := client.Organizations.ListMembers(context.Background(), *org.Login, &github.ListMembersOptions{
				Filter: "2fa_disabled",
			})

			tfaDisabledMemberIDs := make(map[*int64]struct{}, len(tfaDisabledMembers))
			for _, tfaDisabledMember := range tfaDisabledMembers {
				tfaDisabledMemberIDs[tfaDisabledMember.ID] = struct{}{}
			}

			adminMembers, _, err := client.Organizations.ListMembers(context.Background(), *org.Login, &github.ListMembersOptions{
				Role: "admin",
			})
			if err != nil {
				panic(err)
			}

			members, _, err := client.Organizations.ListMembers(context.Background(), *org.Login, &github.ListMembersOptions{
				Role: "member",
			})
			if err != nil {
				panic(err)
			}

			membersTypes := make(map[*string]string, len(adminMembers)+len(members))

			for _, member := range adminMembers {
				membersTypes[member.Login] = "Admin"
			}

			for _, member := range members {
				membersTypes[member.Login] = "Member"
			}

			fmt.Println(*org.Login)

			for _, member := range append(adminMembers, members...) {
				fmt.Printf("  %s - %s\n", *member.Login, membersTypes[member.Login])

				if _, ok := tfaDisabledMemberIDs[member.ID]; ok {
					fmt.Println("    - 2FA Disabled")
				}
			}

			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(githubCmd)

	githubCmd.Flags().StringVarP(&token, "token", "t", "", "The GitHub token")
}
