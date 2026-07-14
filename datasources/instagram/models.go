/*
	Timelinize
	Copyright (c) 2013 Matthew Holt

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published
	by the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package instagram

import (
	"context"
	"io"
	"io/fs"
	"path"
	"time"

	"github.com/timelinize/timelinize/datasources/facebook"
	"github.com/timelinize/timelinize/timeline"
)

type instaPostsIndex []instaPost

type instaPost struct {
	Media []instaMedia `json:"media"`

	// usually for posts with multiple media
	Title             string `json:"title,omitempty"`
	CreationTimestamp int64  `json:"creation_timestamp,omitempty"`

	filename string
}

// allText returns a concatenation of all the titles in the post.
func (p instaPost) allText() string {
	s := facebook.FixString(p.Title)
	for _, m := range p.Media {
		if m.Title != "" {
			if s != "" {
				s += "\n\n"
			}
			s += facebook.FixString(m.Title)
		}
	}
	return s
}

// timestamp returns the timestamp of the post, or if not set,
// the timestamp of the first media item.
func (p instaPost) timestamp() time.Time {
	if p.CreationTimestamp > 0 {
		return time.Unix(p.CreationTimestamp, 0)
	}
	return time.Unix(p.Media[0].CreationTimestamp, 0)
}

type instaMedia struct {
	URI               string         `json:"uri"`
	CreationTimestamp int64          `json:"creation_timestamp"`
	MediaMetadata     instaMediaMeta `json:"media_metadata"`
	Title             string         `json:"title"`
}

func (m instaMedia) timelineItem(fsys fs.FS, owner timeline.Entity) *timeline.Item {
	return &timeline.Item{
		Classification:       timeline.ClassSocial,
		Timestamp:            time.Unix(m.CreationTimestamp, 0),
		Location:             m.location(),
		Owner:                owner,
		IntermediateLocation: m.URI,
		Content: timeline.ItemData{
			Filename: path.Base(m.URI),
			Data: func(_ context.Context) (io.ReadCloser, error) {
				return fsys.Open(m.URI)
			},
		},
	}
}

func (m instaMedia) location() timeline.Location {
	var l timeline.Location
	if len(m.MediaMetadata.PhotoMetadata.ExifData) > 0 {
		l.Latitude = &m.MediaMetadata.PhotoMetadata.ExifData[0].Latitude
		l.Longitude = &m.MediaMetadata.PhotoMetadata.ExifData[0].Longitude
	} else if len(m.MediaMetadata.VideoMetadata.ExifData) > 0 {
		l.Latitude = &m.MediaMetadata.VideoMetadata.ExifData[0].Latitude
		l.Longitude = &m.MediaMetadata.VideoMetadata.ExifData[0].Longitude
	}
	return l
}

type instaMediaMeta struct {
	PhotoMetadata instaPhotoAndVideoMeta `json:"photo_metadata"`
	VideoMetadata instaPhotoAndVideoMeta `json:"video_metadata"`
}

type instaPhotoAndVideoMeta struct {
	ExifData []instaExifData `json:"exif_data"`
}

type instaExifData struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type instaPersonalInformation struct {
	ProfileUser []struct {
		Title        string `json:"title"`
		MediaMapData struct {
			ProfilePhoto struct {
				URI               string `json:"uri"`
				CreationTimestamp int64  `json:"creation_timestamp"`
				Title             string `json:"title"`
			} `json:"Profile Photo"`
		} `json:"media_map_data"`
		StringMapData profileStringMap `json:"string_map_data"`
	} `json:"profile_user"`
}

// profileStringMap is a map rather than a fixed struct because Instagram
// translates these key names based on the exporting account's UI language
// (eg. "Username" becomes "Benutzername" for a German-language account), so
// static json struct tags only match English-language exports and silently
// leave every field empty for anyone else - including the owner's name and
// identity, which then shows up as "(unknown)" throughout the timeline.
type profileStringMap map[string]instaProfileData

// profileFieldAliases holds localized key names Instagram is known to use
// for profile fields, indexed by their canonical (English) name; get()
// always tries the canonical name first. Extend this as more languages are
// confirmed - "Telefonnummer" is inferred from the confirmed German phrase
// "Telefonnummer bestätigt" ("Phone Confirmed") rather than observed
// directly, since accounts without a stored phone number omit the key
// entirely.
var profileFieldAliases = map[string][]string{
	"Email":           {"E-Mail-Adresse"},          // German
	"Phone Number":    {"Telefonnummer"},           // German (inferred, see above)
	"Phone Confirmed": {"Telefonnummer bestätigt"}, // German
	"Username":        {"Benutzername"},            // German
	"Gender":          {"Geschlecht"},              // German
	"Date of birth":   {"Geburtsdatum"},            // German
	"Private Account": {"Privates Konto"},          // German
}

// get returns the field for canonicalKey (its English name), falling back
// to any known localized aliases if the export uses a different language.
func (m profileStringMap) get(canonicalKey string) instaProfileData {
	if v, ok := m[canonicalKey]; ok {
		return v
	}
	for _, alias := range profileFieldAliases[canonicalKey] {
		if v, ok := m[alias]; ok {
			return v
		}
	}
	return instaProfileData{}
}

type instaProfileData struct {
	Href      string `json:"href"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type instaStories struct {
	IgStories []struct {
		URI               string `json:"uri"`
		CreationTimestamp int64  `json:"creation_timestamp"`
		Title             string `json:"title"`
	} `json:"ig_stories"`
}
