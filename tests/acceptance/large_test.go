//go:build large
// +build large

package acceptance

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestLargeFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths: []string{
				"../../features/push.feature",
				"../../features/pull.feature",
				"../../features/remote.feature",
			},
			TestingT: t,
			Tags:     "~@pending",
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run large feature tests")
	}
}
