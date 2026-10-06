package library

import "testing"

func TestRecommendationBlocksRoundTrip(t *testing.T) {
	library, _ := openTestLibrary(t)
	ctx := t.Context()

	if err := library.BlockRecommendation(ctx, BlockTrack, "video1", "Song"); err != nil {
		t.Fatal(err)
	}
	if err := library.BlockRecommendation(ctx, BlockArtist, "UC1", "Artist"); err != nil {
		t.Fatal(err)
	}
	if err := library.BlockRecommendation(ctx, BlockArtist, "UC1", "Renamed"); err != nil {
		t.Fatal(err)
	}

	blocks, err := library.RecommendationBlocks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("blocks = %+v, want 2", blocks)
	}
	names := map[string]string{}
	for _, block := range blocks {
		names[string(block.Kind)+":"+block.ID] = block.Name
	}
	if names["track:video1"] != "Song" || names["artist:UC1"] != "Renamed" {
		t.Errorf("blocks = %+v", blocks)
	}

	if err := library.UnblockRecommendation(ctx, BlockTrack, "video1"); err != nil {
		t.Fatal(err)
	}
	blocks, err = library.RecommendationBlocks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 || blocks[0].ID != "UC1" {
		t.Errorf("after unblock = %+v", blocks)
	}
}

func TestBlockRecommendationRejectsEmptyID(t *testing.T) {
	library, _ := openTestLibrary(t)

	if err := library.BlockRecommendation(t.Context(), BlockArtist, "", "Nobody"); err == nil {
		t.Error("blocked an empty id")
	}
}
