package testsuite

import (
	"testing"

	adapter "github.com/tinode/chat/server/db"
	"github.com/tinode/chat/server/db/common/test_data"
	types "github.com/tinode/chat/server/store/types"
)

// RunReactionsCRUD runs the shared Reactions CRUD tests used by adapters.
func RunReactionsCRUD(t *testing.T, adp adapter.Adapter, td *test_data.TestData) {
	t.Helper()

	// Ensure topic exists (adapter should handle this, but check for sanity)
	if _, err := adp.TopicGet(td.Topics[1].Id); err != nil {
		t.Fatal("TopicGet", err)
	}

	// Use topic[1] which still has messages at this point in the test suite.
	topic := td.Topics[1].Id
	seq := td.Reacts[0].SeqId

	u00 := types.ParseUid(td.Reacts[0].User)

	// No reactions added yet: should return nil map and no error.
	_, err := adp.ReactionGetAll(topic, u00, false, nil)
	if err == nil {
		t.Error("Expected error for nil opts")
	}

	// Save two identical reactions from two users
	r1 := td.Reacts[0]
	r1.MrrId = 1
	if err := adp.ReactionSave(r1); err != nil {
		t.Fatal("ReactionSave 1", err)
	}

	r2 := td.Reacts[1]
	r2.MrrId = 2
	if err := adp.ReactionSave(r2); err != nil {
		t.Fatal("ReactionSave 2", err)
	}

	// Query Since 1 (>= 1). Should return both.
	opts := &types.QueryOpt{Since: 1}
	reacts, err := adp.ReactionGetAll(topic, u00, false, opts)
	if err != nil {
		t.Fatal("ReactionGetAll Since 1", err)
	}
	if len(reacts) == 0 {
		t.Fatal("No reactions returned")
	}
	found := false
	for _, arr := range reacts {
		for _, r := range arr {
			if r.Content == "👍" {
				found = true
				if r.Cnt != 2 {
					t.Error("Expected cnt 2 got", r.Cnt)
				}
				if len(r.Users) != 2 {
					t.Error("Expected 2 users, got", r.Users)
				}
				if r.MrrId != 2 {
					t.Errorf("Expected MrrId 2, got %d", r.MrrId)
				}
			}
		}
	}
	if !found {
		t.Error("expected 👍 reaction")
	}

	// asChan mode: counts and current user's marking
	reactsChan, err := adp.ReactionGetAll(topic, u00, true, opts)
	if err != nil {
		t.Fatal("ReactionGetAll asChan", err)
	}
	rarr := reactsChan[seq]
	if len(rarr) == 0 {
		t.Fatal("No reactions in asChan")
	}
	var gotCnt int
	var gotMarked bool
	for _, r := range rarr {
		if r.Content == "👍" {
			gotCnt = r.Cnt
			if r.MrrId != 2 {
				t.Errorf("asChan Expected MrrId 2, got %d", r.MrrId)
			}
			if len(r.Users) == 1 && r.Users[0] == u00.UserId() {
				gotMarked = true
			}
		}
	}
	if gotCnt != 2 {
		t.Error("asChan cnt expected 2 got", gotCnt)
	}
	if !gotMarked {
		t.Error("current user's reaction not marked")
	}

	// Update reaction by changing u1 reaction to ❤️
	r1b := td.Reacts[2]
	r1b.MrrId = 3 // This won't change the MrrId in DB because of ON DUPLICATE KEY UPDATE content
	if err := adp.ReactionSave(r1b); err != nil {
		t.Fatal("ReactionSave 3", err)
	}
	reacts, err = adp.ReactionGetAll(topic, u00, false, opts)
	if err != nil {
		t.Fatal("ReactionGetAll", err)
	}
	// Expect ❤️ cnt 1 and 👍 cnt 1
	foundPlus := false
	foundHeart := false
	for _, arr := range reacts {
		for _, r := range arr {
			switch r.Content {
			case "👍":
				if r.Cnt != 1 {
					t.Error("Expected 👍 cnt 1 got", r.Cnt)
				}
				if r.MrrId != 1 {
					t.Errorf("Expected 👍 MrrId 1, got %d", r.MrrId)
				}
				foundPlus = true
			case "❤️":
				if r.Cnt != 1 {
					t.Error("Expected ❤️ cnt 1 got", r.Cnt)
				}
				if r.MrrId != 3 {
					t.Errorf("Expected ❤️ MrrId 3, got %d", r.MrrId)
				}
				foundHeart = true
			}
		}
	}
	if !foundPlus || !foundHeart {
		t.Error("expected both 👍 and ❤️ reactions")
	}

	// Test Since > 1. Should return r2 (MrrId 2) only. r1 is MrrId 1.
	// r1b updated r2? No r1b is td.Reacts[2] which is User[1] same as r2.
	// So r2 was updated to Heart. r2 has MrrId 2.
	// r1 (User[0]) is ThumbsUp. MrrId 1.
	// So reacts should contain MrrId 1 (Thumbs) and MrrId 2 (Heart).
	// Query Since 2: Should return MrrId 2 (Heart).
	optsSince2 := &types.QueryOpt{Since: 2}
	reacts, err = adp.ReactionGetAll(topic, u00, false, optsSince2)
	if err != nil {
		t.Fatal("ReactionGetAll Since 2", err)
	}
	// We expect only Heart.
	for _, arr := range reacts {
		for _, r := range arr {
			if r.Content == "👍" {
				t.Error("Did not expect 👍 with Since 2")
			}
			if r.Content == "❤️" {
				if r.Cnt != 1 {
					t.Error("Expected ❤️ cnt 1 got", r.Cnt)
				}
				if r.MrrId != 3 {
					t.Errorf("Since 2 Expected ❤️ MrrId 3, got %d", r.MrrId)
				}
			}
		}
	}

	// Test Before 2. Should return r1 (MrrId 1).
	optsBefore2 := &types.QueryOpt{Before: 2}
	reacts, err = adp.ReactionGetAll(topic, u00, false, optsBefore2)
	if err != nil {
		t.Fatal("ReactionGetAll Before 2", err)
	}
	// We expect only ThumbsUp.
	foundPlus = false
	foundHeart = false
	for _, arr := range reacts {
		for _, r := range arr {
			if r.Content == "👍" {
				foundPlus = true
			}
			if r.Content == "❤️" {
				foundHeart = true
			}
		}
	}
	if !foundPlus {
		t.Error("Expected 👍 with Before 2")
	}
	if foundHeart {
		t.Error("Did not expect ❤️ with Before 2")
	}

	// Delete u0 reaction (MrrId 1)
	if err := adp.ReactionDelete(topic, seq, u00); err != nil {
		t.Fatal("ReactionDelete", err)
	}

	reacts, err = adp.ReactionGetAll(topic, u00, false, opts)
	if err != nil {
		t.Fatal("ReactionGetAll after delete", err)
	}
	if len(reacts) != 1 {
		t.Error("Expected reaction to one message only after delete, got", len(reacts))
	}

	// After delete only ❤️ remains
	for _, arr := range reacts {
		if len(arr) != 1 {
			t.Error("Expected only one reaction type after delete, got", len(arr))
		} else if arr[0].Content != "❤️" {
			t.Error("expected only ❤️ reaction")
		}
	}

	// Check that LIMIT applies to number of rows returned, which corresponds to unique (seqid, content) + mrrid order.
	// The implementation sorts by MrrId DESC.
	// Remaining: MrrId 2 (Heart).

	// Let's add more reactions to test Limit.
	// create new reaction on another seq
	newr := td.Reacts[4]
	newr.MrrId = 4
	if err := adp.ReactionSave(newr); err != nil {
		t.Fatal("ReactionSave new", err)
	}

	// Now we have MrrId 2 (Heart on seq) and MrrId 4 (new on seq2).
	// Limit 1 with Since 1. Order DESC. Should return MrrId 4 (new).
	optsLimit := &types.QueryOpt{Since: 1, Limit: 1}
	reactsLimit, err := adp.ReactionGetAll(topic, u00, false, optsLimit)
	if err != nil {
		t.Fatal("ReactionGetAll with limit", err)
	}
	// Expect only MrrId 4.
	found4 := false
	found2 := false
	for _, arr := range reactsLimit {
		for _, r := range arr {
			if r.Content == "new" {
				found4 = true
			}
			if r.Content == "❤️" {
				found2 = true
			}
		}
	}
	if !found4 {
		t.Error("Expected 'new' reaction (MrrId 4) with Limit 1")
	}
	if found2 {
		t.Error("Did not expect '❤️' reaction (MrrId 2) with Limit 1")
	}

	// Invalid QueryOpt (non-nil but no ranges/since/before) => error
	_, err = adp.ReactionGetAll(topic, u00, false, &types.QueryOpt{})
	if err == nil {
		t.Error("expected error for invalid query options")
	}
}
