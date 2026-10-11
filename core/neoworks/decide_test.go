package neoworks

import (
	"testing"
	"time"
)

func TestDecideFollowsTheMergeRules(t *testing.T) {
	older := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	cases := map[string]struct {
		state situation
		want  action
	}{
		"new local entity":  {situation{hasLocal: true, localHash: "a"}, actionCreateNode},
		"new remote entity": {situation{hasNode: true, remoteHash: "a"}, actionApplyRemote},
		"same content on both sides is only linked": {
			situation{hasLocal: true, localHash: "a", hasNode: true, remoteHash: "a"}, actionLinkOnly,
		},
		"duplicate with a newer node": {
			situation{hasLocal: true, localHash: "a", localUpdatedAt: older, hasNode: true, remoteHash: "b", remoteUpdatedAt: newer}, actionApplyRemote,
		},
		"duplicate with a newer local entity": {
			situation{hasLocal: true, localHash: "a", localUpdatedAt: newer, hasNode: true, remoteHash: "b", remoteUpdatedAt: older}, actionUpdateNode,
		},
		"nothing changed": {
			situation{hasLocal: true, localHash: "a", hasNode: true, remoteHash: "a", linked: true, linkHash: "a"}, actionNothing,
		},
		"only the node moved without a change": {
			situation{hasLocal: true, localHash: "a", hasNode: true, remoteHash: "a", linked: true, linkHash: "a", nodeMoved: true}, actionLinkOnly,
		},
		"only local changed": {
			situation{hasLocal: true, localHash: "b", hasNode: true, remoteHash: "a", linked: true, linkHash: "a"}, actionUpdateNode,
		},
		"only remote changed": {
			situation{hasLocal: true, localHash: "a", hasNode: true, remoteHash: "b", linked: true, linkHash: "a", nodeMoved: true}, actionApplyRemote,
		},
		"both changed and local is newer": {
			situation{hasLocal: true, localHash: "b", localUpdatedAt: newer, hasNode: true, remoteHash: "c", remoteUpdatedAt: older, linked: true, linkHash: "a"}, actionUpdateNode,
		},
		"both changed and remote is newer": {
			situation{hasLocal: true, localHash: "b", localUpdatedAt: older, hasNode: true, remoteHash: "c", remoteUpdatedAt: newer, linked: true, linkHash: "a"}, actionApplyRemote,
		},
		"deleted locally": {
			situation{hasNode: true, remoteHash: "a", linked: true, linkHash: "a"}, actionDeleteNode,
		},
		"deleted locally but edited remotely": {
			situation{hasNode: true, remoteHash: "b", linked: true, linkHash: "a"}, actionApplyRemote,
		},
		"deleted remotely": {
			situation{hasLocal: true, localHash: "a", linked: true, linkHash: "a", tombstone: true}, actionRemoveLocal,
		},
		"deleted remotely but edited locally": {
			situation{hasLocal: true, localHash: "b", linked: true, linkHash: "a", tombstone: true}, actionCreateNode,
		},
		"deleted on both sides": {situation{linked: true, linkHash: "a", tombstone: true}, actionDropLink},
		"node missing without a tombstone is left alone": {
			situation{hasLocal: true, localHash: "a", linked: true, linkHash: "a"}, actionNothing,
		},
		"shared cache keeps its nodes": {
			situation{hasNode: true, remoteHash: "a", linked: true, linkHash: "a", keepsNodes: true}, actionNothing,
		},
		"shared cache recreates a deleted node": {
			situation{hasLocal: true, localHash: "a", linked: true, linkHash: "a", tombstone: true, keepsNodes: true}, actionCreateNode,
		},
		"owned entity repairs a node edited elsewhere": {
			situation{hasLocal: true, localHash: "a", hasNode: true, remoteHash: "b", linked: true, linkHash: "a", owned: true, nodeMoved: true}, actionUpdateNode,
		},
		"owned entity in step with its node": {
			situation{hasLocal: true, localHash: "a", hasNode: true, remoteHash: "a", linked: true, linkHash: "a", owned: true}, actionNothing,
		},
		"owned entity is pushed": {
			situation{hasLocal: true, localHash: "b", hasNode: true, remoteHash: "a", linked: true, linkHash: "a", owned: true}, actionUpdateNode,
		},
		"owned node without an entity is kept": {
			situation{hasNode: true, remoteHash: "a", owned: true}, actionNothing,
		},
	}
	for name, test := range cases {
		if got := decide(test.state); got != test.want {
			t.Errorf("%s: decide = %d, want %d", name, got, test.want)
		}
	}
}
