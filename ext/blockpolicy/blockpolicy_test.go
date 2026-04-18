package blockpolicy

import (
	"testing"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func TestRulesAndNextLevel(t *testing.T) {
	if len(Rules(LevelAggressive)) == 0 {
		t.Fatal("expected aggressive rules")
	}
	if len(Rules(LevelBalanced)) == 0 {
		t.Fatal("expected balanced rules")
	}
	if len(Rules(LevelRelaxed)) == 0 {
		t.Fatal("expected relaxed rules")
	}
	if len(Rules(LevelOff)) != 0 {
		t.Fatal("expected off rules to be empty")
	}

	if NextLevel(LevelAggressive) != LevelBalanced {
		t.Fatal("expected aggressive -> balanced")
	}
	if NextLevel(LevelBalanced) != LevelRelaxed {
		t.Fatal("expected balanced -> relaxed")
	}
	if NextLevel(LevelRelaxed) != LevelOff {
		t.Fatal("expected relaxed -> off")
	}
}

func TestLooksBlocked(t *testing.T) {
	blocked := &phantomjscloud.UserResponseWithMeta{
		UserResponse: phantomjscloud.UserResponse{
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: 200, Content: "Robot or human?"},
			},
		},
	}
	if !LooksBlocked(blocked) {
		t.Fatal("expected challenge content to be blocked")
	}

	ok := &phantomjscloud.UserResponseWithMeta{
		UserResponse: phantomjscloud.UserResponse{
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: 200, Content: "welcome to example"},
			},
		},
	}
	if LooksBlocked(ok) {
		t.Fatal("expected regular page to be non-blocked")
	}
}
