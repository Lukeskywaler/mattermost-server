// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package searchtest

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/mattermost/mattermost-server/v5/utils"

	"github.com/stretchr/testify/assert"
)

const (
	ENGINE_ALL           = "all"
	ENGINE_MYSQL         = "mysql"
	ENGINE_POSTGRES      = "postgres"
	ENGINE_ELASTICSEARCH = "elasticsearch"
)

type SearchTestEngine struct {
	Driver     string
	BeforeTest func(*testing.T, store.Store)
	AfterTest  func(*testing.T, store.Store)
}

type searchTest struct {
	Name string
	Fn   func(*testing.T, store.Store)
	Tags []string
}

func filterTestsByTag(tests []searchTest, tags ...string) []searchTest {
	filteredTests := []searchTest{}
	for _, test := range tests {
		if utils.StringInSlice(ENGINE_ALL, test.Tags) {
			filteredTests = append(filteredTests, test)
			continue
		}
		for _, tag := range tags {
			if utils.StringInSlice(tag, test.Tags) {
				filteredTests = append(filteredTests, test)
				break
			}
		}
	}

	return filteredTests
}

func runTestSearch(t *testing.T, s store.Store, testEngine *SearchTestEngine, tests []searchTest) {
	filteredTests := filterTestsByTag(tests, testEngine.Driver)

	for _, test := range filteredTests {
		if testEngine.BeforeTest != nil {
			testEngine.BeforeTest(t, s)
		}
		t.Run(test.Name, func(t *testing.T) { test.Fn(t, s) })
		if testEngine.AfterTest != nil {
			testEngine.AfterTest(t, s)
		}
	}
}

func makeEmail() string {
	return "success_" + model.NewId() + "@simulator.amazonses.com"
}

func assertUsers(t *testing.T, expected, actual []*model.User) {
	expectedUsernames := make([]string, 0, len(expected))
	for _, user := range expected {
		expectedUsernames = append(expectedUsernames, user.Username)
	}

	actualUsernames := make([]string, 0, len(actual))
	for _, user := range actual {
		actualUsernames = append(actualUsernames, user.Username)
	}

	if assert.Equal(t, expectedUsernames, actualUsernames) {
		assert.Equal(t, expected, actual)
	}
}

func assertUsersMatchInAnyOrder(t *testing.T, expected, actual []*model.User) {
	expectedUsernames := make([]string, 0, len(expected))
	for _, user := range expected {
		expectedUsernames = append(expectedUsernames, user.Username)
	}

	actualUsernames := make([]string, 0, len(actual))
	for _, user := range actual {
		actualUsernames = append(actualUsernames, user.Username)
	}

	if assert.ElementsMatch(t, expectedUsernames, actualUsernames) {
		assert.ElementsMatch(t, expected, actual)
	}
}

func createUser(username, nickname, firstName, lastName string) *model.User {
	user := &model.User{
		Username:  username,
		Password:  username,
		Nickname:  nickname,
		FirstName: firstName,
		LastName:  lastName,
		Email:     makeEmail(),
	}

	return user
}

func addUserToTeamsAndChannels(s store.Store, user *model.User, teamIds []string, channelIds []string) error {
	for _, teamId := range teamIds {
		_, err := s.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: user.Id}, -1)
		if err != nil {
			return err
		}
	}

	for _, channelId := range channelIds {
		_, err := s.Channel().SaveMember(&model.ChannelMember{ChannelId: channelId, UserId: user.Id, NotifyProps: model.GetDefaultChannelNotifyProps()})
		if err != nil {
			return err
		}
	}

	return nil
}

type ByChannelDisplayName model.ChannelList

func (s ByChannelDisplayName) Len() int { return len(s) }
func (s ByChannelDisplayName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByChannelDisplayName) Less(i, j int) bool {
	if s[i].DisplayName != s[j].DisplayName {
		return s[i].DisplayName < s[j].DisplayName
	}

	return s[i].Id < s[j].Id
}
