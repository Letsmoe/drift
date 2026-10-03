package innertube

// Profile is the client identity InnerTube sees. YouTube decides per client which
// endpoints work, whether stream URLs are ciphered and whether PoTokens are required,
// so each kind of request uses the profile that currently works best for it.
//
// Versions and identities track yt-dlp's INNERTUBE_CLIENTS
// (yt_dlp/extractor/youtube/_base.py); update them from there when YouTube rejects one.
type Profile struct {
	Name        string
	ID          int
	Version     string
	UserAgent   string
	Host        string
	DeviceMake  string
	DeviceModel string
	OSName      string
	OSVersion   string
}

// WebRemix is the YouTube Music web client, used for search and browse.
var WebRemix = Profile{
	Name:      "WEB_REMIX",
	ID:        67,
	Version:   "1.20260930.01.00",
	UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36",
	Host:      "music.youtube.com",
}

// VisionOS returns plain, unciphered stream URLs that play without a PoToken.
// It is yt-dlp's default client when no JS runtime is available.
var VisionOS = Profile{
	Name:        "VISIONOS",
	ID:          101,
	Version:     "1.02",
	UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_7_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0 Safari/605.1.15",
	Host:        "www.youtube.com",
	DeviceMake:  "Apple",
	DeviceModel: "RealityDevice17,1",
	OSName:      "visionOS",
	OSVersion:   "26.5.23O471",
}

type requestContext struct {
	Client clientInfo `json:"client"`
}

type clientInfo struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	UserAgent     string `json:"userAgent,omitempty"`
	DeviceMake    string `json:"deviceMake,omitempty"`
	DeviceModel   string `json:"deviceModel,omitempty"`
	OSName        string `json:"osName,omitempty"`
	OSVersion     string `json:"osVersion,omitempty"`
	Language      string `json:"hl"`
	Region        string `json:"gl"`
	VisitorData   string `json:"visitorData,omitempty"`
}

func (profile Profile) origin() string {
	return "https://" + profile.Host
}
