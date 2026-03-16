package main

import "testing"

func TestBuildMediaFromPhotoAndSticker(t *testing.T) {
	width := 640
	height := 480
	fileSize := int64(123)
	msg := tgMessage{
		ID:           42,
		Type:         "message",
		ReplyToID:    7,
		Photo:        "photos/photo.jpg",
		MediaType:    "sticker",
		File:         "stickers/sticker.webp",
		FileName:     "sticker.webp",
		FileSize:     &fileSize,
		StickerEmoji: "??",
		Width:        &width,
		Height:       &height,
	}

	items := buildMedia(msg)
	if len(items) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(items))
	}
	if items[0].Kind != "photo" || items[0].URL != "photos/photo.jpg" {
		t.Fatalf("unexpected photo item: %+v", items[0])
	}
	if items[1].Kind != "gif" || items[1].URL != "stickers/sticker.webp" {
		t.Fatalf("unexpected sticker item: %+v", items[1])
	}
}

func TestBuildMediaSkipsBrokenPreview(t *testing.T) {
	msg := tgMessage{
		ID:        43,
		Type:      "message",
		ReplyToID: 5,
		MediaType: "video_file",
		File:      "video_files/file.webm",
		Thumbnail: "(File not included. Change data exporting settings to download.)",
	}

	items := buildMedia(msg)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PreviewURL != "" {
		t.Fatalf("expected empty preview for broken thumbnail, got %q", items[0].PreviewURL)
	}
}
