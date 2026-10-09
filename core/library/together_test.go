package library

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestTogetherIdentityRoundTrip(t *testing.T) {
	library, _ := openTestLibrary(t)
	ctx := context.Background()

	_, found, err := library.TogetherIdentity(ctx)
	if err != nil || found {
		t.Fatalf("empty library: found=%v err=%v", found, err)
	}
	saved := TogetherIdentity{UserID: "ABCDEFGHJKLM", SigningKey: "secretseed", Name: "Ana", SharePresence: true}
	if err := library.SaveTogetherIdentity(ctx, saved); err != nil {
		t.Fatal(err)
	}
	saved.Name = "Ana B"
	saved.SharePresence = false
	if err := library.SaveTogetherIdentity(ctx, saved); err != nil {
		t.Fatal(err)
	}
	identity, found, err := library.TogetherIdentity(ctx)
	if err != nil || !found || identity != saved {
		t.Fatalf("identity = %v, found=%v err=%v", identity, found, err)
	}
	for _, formatted := range []string{fmt.Sprint(identity), fmt.Sprintf("%#v", identity), fmt.Sprintf("%+v", identity)} {
		if strings.Contains(formatted, "secretseed") {
			t.Fatalf("formatting leaks the key: %s", formatted)
		}
	}
}

func TestTogetherFriends(t *testing.T) {
	library, _ := openTestLibrary(t)
	ctx := context.Background()

	for _, friend := range []TogetherFriend{{UserID: "B", Name: "ben"}, {UserID: "A", Name: "Cleo"}, {UserID: "C", Name: "Ana"}} {
		if err := library.SaveTogetherFriend(ctx, friend); err != nil {
			t.Fatal(err)
		}
	}
	if err := library.SaveTogetherFriend(ctx, TogetherFriend{UserID: "A", Name: "Alma", Trusted: true}); err != nil {
		t.Fatal(err)
	}
	if err := library.RemoveTogetherFriend(ctx, "C"); err != nil {
		t.Fatal(err)
	}
	friends, err := library.TogetherFriends(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(friends) != 2 || friends[0].Name != "Alma" || !friends[0].Trusted || friends[1].Name != "ben" || friends[1].Trusted {
		t.Fatalf("friends = %+v", friends)
	}
	if friends[0].AddedAt.IsZero() {
		t.Fatal("added time missing")
	}
}
